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
			Name:    os.Getenv("KROGER_PROVIDER_NAME"),
			APIRoot: os.Getenv("KROGER_API_ROOT"),
			Paths:   make(map[string]openapi3.PathItem),
		},
		{
			Name:    os.Getenv("GOOGLE_PROVIDER_NAME"),
			APIRoot: os.Getenv("GOOGLE_API_ROOT"),
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
