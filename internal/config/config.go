package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	KrogerClientID     string
	KrogerClientSecret string
	KrogerRedirectURI  string
	KrogerScopes       string
	DBConnectionString string
	OryActionsSecret   string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	return &Config{
		KrogerClientID:     os.Getenv("KROGER_CLIENT_ID"),
		KrogerClientSecret: os.Getenv("KROGER_CLIENT_SECRET"),
		KrogerRedirectURI:  os.Getenv("KROGER_REDIRECT_URI"),
		KrogerScopes:       os.Getenv("KROGER_SCOPES"),
		DBConnectionString: os.Getenv("DB_CONNECTION_STRING"),
		OryActionsSecret:   os.Getenv("ORY_ACTIONS_SECRET"),
	}
}
