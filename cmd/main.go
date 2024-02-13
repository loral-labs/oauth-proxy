package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"lorallabs.com/oauth-server/cmd/utils"
	"lorallabs.com/oauth-server/internal/config"
	"lorallabs.com/oauth-server/internal/oauthserver"
	"lorallabs.com/oauth-server/internal/store"
	"lorallabs.com/oauth-server/internal/types"

	ory "github.com/ory/client-go"
	"github.com/rs/cors"
)

func main() {
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
		// Enable Debugging for testing, consider disabling in production
		// Debug: true,
	})

	// Use this context to access Ory APIs which require an Ory API Key.
	var ctx = context.WithValue(context.Background(), ory.ContextAccessToken, os.Getenv("ORY_API_KEY"))
	oryClient := oauthserver.NewOryClient(ctx)

	ctx = context.WithValue(ctx, types.OryClientKey, oryClient)
	ctx = context.WithValue(ctx, types.ConfigKey, config)
	ctx = context.WithValue(ctx, types.StoreKey, store)

	// oryClient.ListClients("")
	// oryClient.AddScope("aca314fc-8db0-4840-857c-99343e7d40c7", "ji")

	// Register a catch-all handler
	handler := http.NewServeMux()

	// Load and register dynamic endpoints
	utils.RegisterDynamicEndpoints(ctx, handler)

	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		println("Catch-all handler")
		http.NotFound(w, r)
	})

	// Wrap the main handler with CORS middleware
	finalHandler := corsWrapper.Handler(handler)

	log.Default().Println("Server started on :8081")
	http.ListenAndServe(":8081", finalHandler)
}
