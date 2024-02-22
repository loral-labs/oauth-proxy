package utils

import (
	"context"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"lorallabs.com/oauth-server/internal/config"
	"lorallabs.com/oauth-server/internal/oauth"
	"lorallabs.com/oauth-server/internal/oauthserver"
	"lorallabs.com/oauth-server/internal/store"
	"lorallabs.com/oauth-server/internal/types"
)

// AuthMiddleware checks if the request is authenticated
func AuthMiddleware(ctx context.Context, next http.HandlerFunc, provider string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Accept expired or out-of-scope tokens for testing purposes
		laxAuthFlag := ctx.Value(types.LaxAuthFlag).(*bool)

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
		log.Default().Printf("Ory User ID: %s", oryUserID)
		userID, err := uuid.Parse(oryUserID)
		if err != nil {
			http.Error(w, "Invalid User ID", http.StatusUnauthorized)
			return
		}

		// Check if token is active and in scope
		if !introspected.Active && !*laxAuthFlag {
			http.Error(w, "Invalid Token or Out Of Scope", http.StatusUnauthorized)
			return
		}

		// Create a new context with the bearer token
		ogContext := r.Context()
		ctxWithToken := context.WithValue(ogContext, types.BearerTokenKey, token)
		ctxWithToken = context.WithValue(ctxWithToken, types.OryUserIDKey, userID)

		// If authenticated, call the next handler
		next(w, r.WithContext(ctxWithToken))
	}
}

func RegisterDynamicEndpoints(ctx context.Context, handler *mux.Router) {
	config := ctx.Value(types.ConfigKey).(*config.Config)
	store := ctx.Value(types.StoreKey).(*store.Store)

	// master directory of providers
	allProviders := config.Providers

	for _, provider := range allProviders {
		provider := provider // create a new variable to avoid improper closure

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
		log.Default().Printf("Authentication Registered %s", "/"+provider.Name+"/auth")
		authHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Default().Printf("Authenticating %s", "/"+provider.Name+"/auth")
			oauthHandler.HandleAuth(provider.Name, w, r)
		})
		handler.Handle("/"+provider.Name+"/auth", AuthMiddleware(ctx, authHandler, provider.Name))

		// search for endpoints
		// handler.Handle("/search", AuthMiddleware(ctx, HandleSearch, provider.Name))
		// handler.Handle("/store", AuthMiddleware(ctx, HandleStore, provider.Name))

		// handle oauth callback from provider
		log.Default().Printf("Auth Callback Registered %s", "/"+provider.Name+"/auth/callback")
		handler.HandleFunc("/"+provider.Name+"/auth/callback", func(w http.ResponseWriter, r *http.Request) {
			log.Default().Printf("Auth Callback Hit %s", "/"+provider.Name+"/auth/callback")
			oauthHandler.HandleCallback(provider.Name, w, r)
		})

		// Provider function endpoints
		for path, pathItem := range provider.Paths {
			// Create new variables to avoid improper closure
			path := path
			pathItem := pathItem

			for method, operation := range pathItem.Operations() {

				handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					log.Default().Printf("%s%s hit", provider.Name, path)

					if r.Method != method {
						http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
						return
					}

					// Parse path parameters using Gorilla Mux
					vars := mux.Vars(r)
					truePath := path // truePath is the actual path to the provider's API — for now we mirror real routes

					// Replace placeholders in the path with actual parameter values
					for paramName, paramValue := range vars {
						placeholder := "{" + paramName + "}"
						truePath = strings.Replace(truePath, placeholder, paramValue, 1)
					}

					// Get oryUserID from context
					oryUserID := r.Context().Value(types.OryUserIDKey).(uuid.UUID)

					// Get the provider-specific bearer token from the user id
					bearerToken := oauthHandler.HandleGetToken(provider.Name, oryUserID)

					// Construct a request to the true path
					params := r.URL.Query()
					for _, parameter := range operation.Parameters {
						p := *parameter.Value
						if p.Required && p.In == "query" {
							if _, ok := params[p.Name]; !ok {
								http.Error(w, "Missing required parameter: "+p.Name, http.StatusBadRequest)
								return
							}
						}
						if p.Required && p.In == "path" {
							if _, ok := vars[p.Name]; !ok {
								http.Error(w, "Missing required parameter: "+p.Name, http.StatusBadRequest)
								return
							}
						}
					}

					// check if params is length 0
					if len(params) != 0 {
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

					log.Default().Printf("Request: %v\n", req.URL.String())
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
				handler.Handle("/"+provider.Name+"/v1"+path, AuthMiddleware(ctx, handlerFunc, provider.Name)).Methods(method)
			}
		}
	}
}
