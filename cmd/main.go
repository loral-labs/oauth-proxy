package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"lorallabs.com/oauth-server/internal/config"
	"lorallabs.com/oauth-server/internal/oauth"
	"lorallabs.com/oauth-server/internal/store"
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

func registerDynamicEndpoints(store *store.Store) {
	// Define a struct to match the expected format of each endpoint in the JSON
	type Endpoint struct {
		ID         string            `json:"id"`
		Provider   string            `json:"provider"`
		LoralPath  string            `json:"path"`
		TruePath   string            `json:"truePath"`
		HttpMethod string            `json:"httpMethod"`
		Parameters map[string]string `json:"parameters"`
		// Add other fields as necessary
	}

	// Define a struct to hold the array of endpoints
	var endpoints struct {
		Endpoints []Endpoint `json:"endpoints"`
	}

	// Read the JSON file (adjust the path to where your JSON file is located)
	configFile, err := os.Open("internal/apps/kroger/loral_manual.json")
	if err != nil {
		log.Fatal(err)
	}
	defer configFile.Close()

	// Parse the JSON file into the endpoints struct
	if err := json.NewDecoder(configFile).Decode(&endpoints); err != nil {
		log.Fatal(err)
	}

	// Register each endpoint
	for _, endpoint := range endpoints.Endpoints {
		http.HandleFunc("/"+endpoint.LoralPath, func(w http.ResponseWriter, r *http.Request) {
			// Expect the request to match the expected HTTP method
			if r.Method != endpoint.HttpMethod {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			params := r.URL.Query()
			for key, value := range endpoint.Parameters {
				if params.Get(key) != value {
					http.Error(w, "Invalid parameters", http.StatusBadRequest)
					return
				}
			}

			// params must include userId
			userId := r.URL.Query().Get("userId")
			if userId == "" {
				http.Error(w, "Missing userId", http.StatusBadRequest)
				return
			}
			// get token from userId
			token, err := store.GetTokenByClientID(userId, endpoint.Provider)
			if err != nil {
				http.Error(w, "Token not found", http.StatusNotFound)
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
			req, err := http.NewRequest(r.Method, truePath, r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Add authentication headers, if necessary
			req.Header.Add("Authorization", "Bearer "+token.AccessToken)

			// Forward the request to the true path
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
