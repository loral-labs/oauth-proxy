package main

import (
	"log"
	"net/http"

	"lorallabs.com/oauth-server/cmd/utils"
	"lorallabs.com/oauth-server/internal/config"
	"lorallabs.com/oauth-server/internal/oauth"
	"lorallabs.com/oauth-server/internal/store"
)

func main() {
	config := config.LoadConfig()
	store, err := store.NewStore(config.DBConnectionString)
	if err != nil {
		log.Fatal(err)
	}

	oauthHandler := oauth.NewOAuthHandler(config, store)

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
	utils.RegisterDynamicEndpoints(store)

	log.Default().Println("Server started on :8081")
	http.ListenAndServe(":8081", nil)
}
