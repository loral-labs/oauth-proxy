package main

import (
	"log"
	"lorallabs.com/oauth-server/internal/config"
	"lorallabs.com/oauth-server/internal/oauth"
	"lorallabs.com/oauth-server/internal/store"
	"net/http"
)

func main() {
	cfg := config.LoadConfig()
	db, err := store.NewStore(cfg.DBConnectionString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	oAuthHandler := oauth.NewOAuthHandler(cfg, db)

	http.HandleFunc("/login", oAuthHandler.HandleLogin)
	http.HandleFunc("/auth/callback", oAuthHandler.HandleCallback)

	log.Println("Server starting on :8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
