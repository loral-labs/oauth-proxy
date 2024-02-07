package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	ClientID           string
	ClientSecret       string
	RedirectURI        string
	Scopes             string
	DBConnectionString string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	return &Config{
		ClientID:           os.Getenv("KROGER_CLIENT_ID"),
		ClientSecret:       os.Getenv("KROGER_CLIENT_SECRET"),
		RedirectURI:        os.Getenv("KROGER_REDIRECT_URI"),
		Scopes:             os.Getenv("KROGER_SCOPES"),
		DBConnectionString: os.Getenv("DB_CONNECTION_STRING"),
	}
}
