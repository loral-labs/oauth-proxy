package main

import (
	"log"

	"lorallabs.com/oauth-server/internal/config"
	"lorallabs.com/oauth-server/internal/store"
	schema "lorallabs.com/oauth-server/pkg/db"
)

func main() {
	c := config.LoadConfig()

	s, err := store.NewStore(c.DBConnectionString)
	if err != nil {
		log.Fatal(err)
	}

	s.DB.Create(&schema.Provider{
		Name: "kroger",
	})
}
