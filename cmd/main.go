package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"lorallabs.com/oauth-server/cmd/utils"
	"lorallabs.com/oauth-server/internal/config"
	"lorallabs.com/oauth-server/internal/oauth"
	"lorallabs.com/oauth-server/internal/oauthserver"
	"lorallabs.com/oauth-server/internal/store"

	ory "github.com/ory/client-go"
)

func main() {
	config := config.LoadConfig()
	store, err := store.NewStore(config.DBConnectionString)
	if err != nil {
		log.Fatal(err)
	}

	oauthHandler := oauth.NewOAuthHandler(config, store)

	// Use this context to access Ory APIs which require an Ory API Key.
	var oryAuthedContext = context.WithValue(context.Background(), ory.ContextAccessToken, os.Getenv("ORY_API_KEY"))
	oryClient := oauthserver.NewOryClient(oryAuthedContext)
	// save OryClient to the context
	ctx := context.WithValue(context.Background(), "OryClient", oryClient)

	oryClient.ListClients("")
	// oryClient.AddScope("aca314fc-8db0-4840-857c-99343e7d40c7", "ji")

	http.HandleFunc("/auth/", func(w http.ResponseWriter, r *http.Request) {
		providerName := r.URL.Path[len("/auth/"):]
		oauthHandler.HandleAuth(providerName, w, r)
	})

	http.HandleFunc("/auth/callback/", func(w http.ResponseWriter, r *http.Request) {
		providerName := r.URL.Path[len("/auth/callback/"):]
		oauthHandler.HandleCallback(providerName, w, r)
	})

	http.HandleFunc("/auth/token/", func(w http.ResponseWriter, r *http.Request) {
		providerName := r.URL.Query().Get("provider")
		userID := r.URL.Query().Get("user_id")
		oauthHandler.HandleGetToken(providerName, userID, w, r)
	})

	// Load and register dynamic endpoints
	utils.RegisterDynamicEndpoints(ctx, store)

	// Register a catch-all handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	log.Default().Println("Server started on :8081")
	http.ListenAndServe(":8081", nil)
}
