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
)

func main() {
	config := config.LoadConfig()
	store, err := store.NewStore(config.DBConnectionString)
	if err != nil {
		log.Fatal(err)
	}

	// Use this context to access Ory APIs which require an Ory API Key.
	var ctx = context.WithValue(context.Background(), ory.ContextAccessToken, os.Getenv("ORY_API_KEY"))
	oryClient := oauthserver.NewOryClient(ctx)

	ctx = context.WithValue(ctx, types.OryClientKey, oryClient)
	ctx = context.WithValue(ctx, types.ConfigKey, config)
	ctx = context.WithValue(ctx, types.StoreKey, store)

	// oryClient.ListClients("")

	// oryClient.AddScope("aca314fc-8db0-4840-857c-99343e7d40c7", "ji")

	// Load and register dynamic endpoints
	utils.RegisterDynamicEndpoints(ctx)

	// Register a catch-all handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	log.Default().Println("Server started on :8081")
	http.ListenAndServe(":8081", nil)
}
