package main

import (
	"log"
	"net/http"

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

	log.Default().Println("Server started on :8081")
	http.ListenAndServe(":8081", nil)
}
