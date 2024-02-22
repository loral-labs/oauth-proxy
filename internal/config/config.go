package config

import (
	"log"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/joho/godotenv"
)

type Provider struct {
	Name    string
	APIRoot string
	Paths   map[string]openapi3.PathItem
}

type Config struct {
	Providers []Provider

	KrogerClientID     string
	KrogerClientSecret string
	KrogerRedirectURI  string
	KrogerScopes       string

	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURI  string
	GoogleScopes       string

	DBConnectionString string
	OryActionsSecret   string
}

func LoadConfig() *Config {
	// load from .env file if it exists, otherwise in prod enviroment
	err := godotenv.Load()
	if err != nil {
		log.Default().Printf("Error loading .env file: %v\n", err)
	}

	providers := []Provider{
		{
			Name:    "kroger",
			APIRoot: "https://api.kroger.com",
			Paths:   make(map[string]openapi3.PathItem),
		},
		{
			Name:    "google",
			APIRoot: "https://www.googleapis.com",
			Paths:   make(map[string]openapi3.PathItem),
		},
	}

	return &Config{
		Providers: providers,

		KrogerClientID:     os.Getenv("KROGER_CLIENT_ID"),
		KrogerClientSecret: os.Getenv("KROGER_CLIENT_SECRET"),
		KrogerRedirectURI:  os.Getenv("KROGER_REDIRECT_URI"),
		KrogerScopes:       os.Getenv("KROGER_SCOPES"),

		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectURI:  os.Getenv("GOOGLE_REDIRECT_URI"),
		GoogleScopes:       os.Getenv("GOOGLE_SCOPES"),

		DBConnectionString: os.Getenv("DB_CONNECTION_STRING"),
		OryActionsSecret:   os.Getenv("ORY_ACTIONS_SECRET"),
	}
}
