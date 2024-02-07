package main

import (
	"log"
	"net/http"

	"lorallabs.com/oauth-server/internal/config"
	"lorallabs.com/oauth-server/internal/oauth"
)

func main() {
	config := config.LoadConfig()

	providers := oauth.InitializeProviders(config)

	http.HandleFunc("/auth/", func(w http.ResponseWriter, r *http.Request) {
		// Extract provider from the URL
		providerName := r.URL.Path[len("/auth/"):]

		// Initiate OAuth flow
		oauth.HandleAuth(providers, providerName, w, r)
	})

	http.HandleFunc("/auth/callback/", func(w http.ResponseWriter, r *http.Request) {
		// Extract provider from the URL
		providerName := r.URL.Path[len("/auth/callback/"):]

		// Handle OAuth callback
		oauth.HandleCallback(providers, providerName, w, r)
	})

	log.Default().Println("Server started on :8081")
	http.ListenAndServe(":8081", nil)
}
