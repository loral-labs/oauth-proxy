package main

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"os"

	"lorallabs.com/oauth-server/cmd/utils"
	"lorallabs.com/oauth-server/internal/config"
	"lorallabs.com/oauth-server/internal/oauthserver"
	"lorallabs.com/oauth-server/internal/store"
	"lorallabs.com/oauth-server/internal/types"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	ory "github.com/ory/client-go"
	"github.com/rs/cors"
	schema "lorallabs.com/oauth-server/pkg/db"
)

func main() {
	laxAuthFlag := flag.Bool("lax_auth", false, "accept expired or out-of-scope tokens for testing purposes")
	flag.Parse()

	config := config.LoadConfig()
	store, err := store.NewStore(config.DBConnectionString)
	if err != nil {
		log.Fatal(err)
	}

	// Setup CORS
	corsWrapper := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // or use "*" to allow any origin
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	})

	// Use this context to access Ory APIs which require an Ory API Key.
	var ctx = context.WithValue(context.Background(), ory.ContextAccessToken, os.Getenv("ORY_API_KEY"))

	ctx = context.WithValue(ctx, types.ConfigKey, config)
	ctx = context.WithValue(ctx, types.StoreKey, store)
	ctx = context.WithValue(ctx, types.LaxAuthFlag, laxAuthFlag)
	oryClient := oauthserver.NewOryClient(ctx)
	ctx = context.WithValue(ctx, types.OryClientKey, oryClient)

	handler := mux.NewRouter()

	handler.HandleFunc("/auth/introspect", oryClient.ListAppsHandler).Methods("GET")
	handler.HandleFunc("/ory/actions/newUserCallback", func(w http.ResponseWriter, r *http.Request) {
		secret := r.Header.Get("X-Secret")
		if secret != config.OryActionsSecret {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// read the body as a json
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		type UserCallback struct {
			UserID string `json:"userId"`
			Email  string `json:"email"`
		}
		var callbackData UserCallback

		err = json.Unmarshal(body, &callbackData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// create a new user in the database
		uuid, err := uuid.Parse(callbackData.UserID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		newUser := schema.User{
			ID:       uuid,
			Username: callbackData.Email,
			Email:    callbackData.Email,
		}
		log.Printf("Creating new user: %v", newUser)
		err = store.DB.Create(&newUser).Error
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("Created new user: %v", newUser)

		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}).Methods("POST")

	// Load and register dynamic endpoints
	oryClient.RegisterOAuthServerHandlers(handler)
	utils.RegisterDynamicEndpoints(ctx, handler)

	// Register a catch-all 404 handler
	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		println("Catch-all handler")
		http.NotFound(w, r)
	})

	// Wrap the main handler with CORS middleware
	finalHandler := corsWrapper.Handler(handler)

	log.Default().Println("Server started on :8081")
	http.ListenAndServe(":8081", finalHandler)
}
