package utils

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"lorallabs.com/oauth-server/internal/config"
	"lorallabs.com/oauth-server/internal/oauth"
	"lorallabs.com/oauth-server/internal/oauthserver"
	"lorallabs.com/oauth-server/internal/store"
	"lorallabs.com/oauth-server/internal/types"
)

type Config struct {
	Provider  string              `json:"provider"`
	APIRoot   string              `json:"apiroot"`
	Endpoints map[string]Endpoint `json:"endpoints"`
}

type Parameter struct {
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Location    string `json:"location"`
	Format      string `json:"format,omitempty"` // Optional, only for integer parameters
}

type Endpoint struct {
	ID          string               `json:"id"`
	LoralPath   string               `json:"loralPath"`
	TruePath    string               `json:"truePath"`
	HttpMethod  string               `json:"httpMethod"`
	Description string               `json:"description,omitempty"` // Optional
	Parameters  map[string]Parameter `json:"parameters"`
	Response    interface{}          `json:"response,omitempty"` // Optional, adjust as needed
	Request     interface{}          `json:"request,omitempty"`  // Optional, adjust as needed
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

func RegisterDynamicEndpoints(ctx context.Context, handler *http.ServeMux) {
	config := ctx.Value(types.ConfigKey).(*config.Config)
	store := ctx.Value(types.StoreKey).(*store.Store)

	// Read the JSON file (adjust the path to where your JSON file is located)
	configFile, err := os.Open("internal/apps/kroger/loral_manual.json")
	if err != nil {
		log.Fatal(err)
	}
	defer configFile.Close()

	// Parse the JSON file into the Config struct
	var provider Config
	if err := json.NewDecoder(configFile).Decode(&provider); err != nil {
		log.Fatal(err)
	}

	oauthHandler := oauth.NewOAuthHandler(config, store)
	authHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		oauthHandler.HandleAuth(provider.Provider, w, r)
	})
	handler.Handle("/"+provider.Provider+"/auth/", AuthMiddleware(ctx, authHandler, provider.Provider))

	handler.HandleFunc("/"+provider.Provider+"/auth/callback/", func(w http.ResponseWriter, r *http.Request) {
		oauthHandler.HandleCallback(provider.Provider, w, r)
	})

	// Provider function endpoints
	for _, endpoint := range provider.Endpoints {
		endpoint := endpoint // Create a new variable to avoid improper closure

		handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Default().Printf("%s/%s hit", provider.Provider, endpoint.LoralPath)

			if r.Method != endpoint.HttpMethod {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			// Get oryUserID from context
			oryUserID := r.Context().Value(types.OryUserIDKey).(string)

			// Get the provider-specific bearer token from the user id
			bearerToken := oauthHandler.HandleGetToken(provider.Provider, oryUserID)

			// Construct a request to the true path
			truePath := endpoint.TruePath
			params := r.URL.Query()
			if len(params) > 0 {
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
		handler.Handle("/"+provider.Provider+"/"+endpoint.LoralPath, AuthMiddleware(ctx, handlerFunc, provider.Provider))
	}
}
