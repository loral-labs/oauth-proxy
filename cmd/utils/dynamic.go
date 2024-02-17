package utils

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"lorallabs.com/oauth-server/internal/config"
	"lorallabs.com/oauth-server/internal/oauth"
	"lorallabs.com/oauth-server/internal/oauthserver"
	"lorallabs.com/oauth-server/internal/store"
	"lorallabs.com/oauth-server/internal/types"
)

type Provider struct {
	Name    string
	APIRoot string
	Paths   map[string]openapi3.PathItem
}

// Struct to hold the environment variables
type EnvConfig struct {
	ESURL          string
	MasterUsername string
	MasterPassword string
	S3Bucket       string
}

var env EnvConfig

func init() {
	// Load environment variables from .env file
	os.Clearenv()
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Read environment variables
	env.ESURL = os.Getenv("LORAL_ES_DOMAIN")
	env.MasterUsername = os.Getenv("LORAL_ES_DOMAIN_USER")
	env.MasterPassword = os.Getenv("LORAL_ES_DOMAIN_PSWD")
}

// AuthMiddleware checks if the request is authenticated
func AuthMiddleware(ctx context.Context, next http.HandlerFunc, provider string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Split the header on the space to separate "Bearer" from the "<token>"
		authHeader := r.Header.Get("Authorization")
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Unauthorized - Invalid token format", http.StatusUnauthorized)
			return
		}
		token := parts[1]

		// get oryclient from context
		o := ctx.Value(types.OryClientKey).(*oauthserver.OryClient)
		introspected := o.IntrospectToken(token, provider)

		oryUserID := introspected.GetSub()
		userID, err := uuid.Parse(oryUserID)
		if err != nil {
			http.Error(w, "Invalid User ID", http.StatusUnauthorized)
			return
		}

		// Check if token is active and in scope
		if !introspected.Active {
			http.Error(w, "Invalid Token or Out Of Scope", http.StatusUnauthorized)
			return
		}

		// Create a new context with the bearer token
		ctxWithToken := context.WithValue(ctx, types.BearerTokenKey, token)
		ctxWithToken = context.WithValue(ctxWithToken, types.OryUserIDKey, userID)

		// If authenticated, call the next handler
		next(w, r.WithContext(ctxWithToken))
	}
}

func RegisterDynamicEndpoints(ctx context.Context, handler *mux.Router) {
	config := ctx.Value(types.ConfigKey).(*config.Config)
	store := ctx.Value(types.StoreKey).(*store.Store)

	// create an instance of the Provider struct
	provider := Provider{
		Name:    "kroger",
		APIRoot: "https://api.kroger.com",
		Paths:   make(map[string]openapi3.PathItem),
	}
	dirPath := "internal/apps/" + provider.Name
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		log.Fatalf("Failed to read directory: %v", err)
	}

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	// parse all files for openapi3 paths
	for _, file := range files {
		if !file.IsDir() {
			filePath := filepath.Join(dirPath, file.Name())
			doc, err := loader.LoadFromFile(filePath)
			if err != nil {
				log.Fatalf("Failed to load OpenAPI document: %v", err)
			}

			if err := doc.Validate(ctx); err != nil {
				log.Fatalf("Failed to validate OpenAPI document: %v", err)
			}

			// add to allPaths
			paths := *doc.Paths
			for path, pathItem := range paths.Map() {
				provider.Paths[path] = *pathItem
			}
		}
	}

	// auth to the provider
	oauthHandler := oauth.NewOAuthHandler(config, store)
	authHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		oauthHandler.HandleAuth(provider.Name, w, r)
	})
	handler.Handle("/"+provider.Provider+"/auth/", AuthMiddleware(ctx, authHandler, provider.Provider))

	// search for endpoints
	handler.Handle("/search", AuthMiddleware(ctx, HandleSearch, provider.Provider))
	handler.Handle("/store", AuthMiddleware(ctx, HandleStore, provider.Provider))

	// handle oauth callback from provider
	handler.HandleFunc("/"+provider.Name+"/auth/callback/", func(w http.ResponseWriter, r *http.Request) {
		oauthHandler.HandleCallback(provider.Name, w, r)
	})

	// Provider function endpoints
	for path, pathItem := range provider.Paths {
		// Create new variables to avoid improper closure
		path := path
		pathItem := pathItem

		// FIX ME - assumes only one method per path
		var method string
		var operation *openapi3.Operation
		for temp_method, temp_op := range pathItem.Operations() {
			operation = temp_op
			method = temp_method
			break
		}

		handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Default().Printf("%s%s hit", provider.Name, path)

			if r.Method != method {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			// Parse path parameters using Gorilla Mux
			fmt.Println(r.URL)
			vars := mux.Vars(r) // NOT GETTING ANYTHING
			truePath := path

			fmt.Println(vars)
			// Replace placeholders in the path with actual parameter values
			for paramName, paramValue := range vars {
				fmt.Println(paramName, paramValue)
				placeholder := "{" + paramName + "}"
				truePath = strings.Replace(truePath, placeholder, paramValue, 1)
			}

			// Get oryUserID from context
			oryUserID := r.Context().Value(types.OryUserIDKey).(uuid.UUID)

			// Get the provider-specific bearer token from the user id
			bearerToken := oauthHandler.HandleGetToken(provider.Name, oryUserID)

			// Construct a request to the true path
			// truePath := path
			params := r.URL.Query()
			if len(params) > 0 {
				// verify that required parameters are present
				for _, parameter := range operation.Parameters {
					p := *parameter.Value
					if p.Required && p.In == "query" {
						if _, ok := params[p.Name]; !ok {
							http.Error(w, "Missing required parameter: "+p.Name, http.StatusBadRequest)
							return
						}
					}
				}

				truePath += "?"
				for key, value := range params {
					truePath += key + "=" + value[0] + "&"
				}
				truePath = truePath[:len(truePath)-1]
			}
			req, err := http.NewRequest(r.Method, provider.APIRoot+truePath, r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Add authentication headers, if necessary
			req.Header.Add("Authorization", "Bearer "+bearerToken)

			// Forward the request to the true path
			log.Default().Printf("Request: %v\n", req)
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer resp.Body.Close()

			// Copy the response body to the original response writer - just the body, not the headers
			w.WriteHeader(resp.StatusCode)
			_, err = io.Copy(w, resp.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		})

		// Wrap in AuthMiddleware
		log.Default().Printf("Registering %s, %s", method, "/"+provider.Name+path)
		handler.Handle("/"+provider.Name+path, AuthMiddleware(ctx, handlerFunc, provider.Name)).Methods(method)
	}
}
