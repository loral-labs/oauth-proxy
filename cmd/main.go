package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"lorallabs.com/oauth-server/internal/config"
	"lorallabs.com/oauth-server/internal/oauth"
	"lorallabs.com/oauth-server/internal/store"
	schema "lorallabs.com/oauth-server/pkg/db"
)

func main() {
	config := config.LoadConfig()
	store, err := store.NewStore(config.DBConnectionString)
	if err != nil {
		log.Fatal(err)
	}

	oauthHandler := oauth.NewOAuthHandler(config, store)

	http.HandleFunc("/auth/", func(w http.ResponseWriter, r *http.Request) {
		providerName := r.URL.Path[len("/auth/"):]
		oauthHandler.HandleAuth(providerName, w, r)
	})

	http.HandleFunc("/auth/callback/", func(w http.ResponseWriter, r *http.Request) {
		providerName := r.URL.Path[len("/auth/callback/"):]
		oauthHandler.HandleCallback(providerName, w, r)
	})

	http.HandleFunc("/auth/token/", func(w http.ResponseWriter, r *http.Request) {
		providerName := r.URL.Query().Get("provider")
		userID := r.URL.Query().Get("user_id")
		oauthHandler.HandleGetToken(providerName, userID, w, r)
	})

	// Load and register dynamic endpoints
	registerDynamicEndpoints(store)

	log.Default().Println("Server started on :8081")
	http.ListenAndServe(":8081", nil)
}

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

func registerDynamicEndpoints(store *store.Store) {
	// Read the JSON file (adjust the path to where your JSON file is located)
	configFile, err := os.Open("internal/apps/kroger/loral_manual.json")
	if err != nil {
		log.Fatal(err)
	}
	defer configFile.Close()

	// Parse the JSON file into the Config struct
	var config Config
	if err := json.NewDecoder(configFile).Decode(&config); err != nil {
		log.Fatal(err)
	}

	for _, endpoint := range config.Endpoints {
		endpoint := endpoint // Create a new variable to avoid improper closure

		http.HandleFunc("/"+config.Provider+"/"+endpoint.LoralPath, func(w http.ResponseWriter, r *http.Request) {
			log.Default().Printf("%s/%s hit", config.Provider, endpoint.LoralPath)

			if r.Method != endpoint.HttpMethod {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			params := r.URL.Query()
			// return descriptive error if userId is missing
			userId := params.Get("userId")
			if userId == "" {
				http.Error(w, "Missing userId, a stable and unique client identifier is required from you", http.StatusBadRequest)
				return
			}
			// check that all required parameters are present
			for key, value := range endpoint.Parameters {
				if value.Required && params.Get(key) == "" {
					http.Error(w, "Missing required parameter: "+key, http.StatusBadRequest)
					return
				}
			}

			// get token from userId
			var clientRecord schema.Client
			err := store.DB.Where("identifier = ?", params.Get("userId")).First(&clientRecord).Error
			if err != nil {
				http.Error(w, "Client not found", http.StatusNotFound)
				return
			}
			// make network request to get bearer token
			tokenURL := fmt.Sprintf("http://%s/auth/token/?provider=%s&user_id=%s", r.Host, config.Provider, userId)
			tokenReq, err := http.NewRequest(http.MethodGet, tokenURL, nil)
			if err != nil {
				http.Error(w, "Failed to create request for token: "+err.Error(), http.StatusInternalServerError)
				return
			}

			tokenResp, err := http.DefaultClient.Do(tokenReq)
			if err != nil {
				http.Error(w, "Failed to get token: "+err.Error(), http.StatusInternalServerError)
				return
			}
			defer tokenResp.Body.Close()

			if tokenResp.StatusCode != http.StatusOK {
				http.Error(w, "Failed to get token, status code: "+strconv.Itoa(tokenResp.StatusCode), http.StatusInternalServerError)
				return
			}

			var tokenData struct {
				BearerToken string `json:"bearer_token"`
			}
			if err := json.NewDecoder(tokenResp.Body).Decode(&tokenData); err != nil {
				http.Error(w, "Failed to decode token response: "+err.Error(), http.StatusInternalServerError)
				return
			}

			// Construct a request to the true path
			truePath := endpoint.TruePath
			if len(params) > 0 {
				truePath += "?"
				for key, value := range params {
					truePath += key + "=" + value[0] + "&"
				}
				truePath = truePath[:len(truePath)-1]
			}
			req, err := http.NewRequest(r.Method, config.APIRoot+truePath, r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Add authentication headers, if necessary
			req.Header.Add("Authorization", "Bearer "+tokenData.BearerToken)

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
	}
}
