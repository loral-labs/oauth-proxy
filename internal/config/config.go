package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
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
	// load from .env file if it exists, otherwise in prod enviroment
	err := godotenv.Load()
	if err != nil {
		log.Default().Printf("Error loading .env file: %v\n", err)
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
