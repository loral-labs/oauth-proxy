package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/joho/godotenv"
)

// Struct to hold the environment variables
type EnvConfig struct {
	ESURL          string
	MasterUsername string
	MasterPassword string
}

var env EnvConfig

func init() {
	// Load environment variables from .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Read environment variables
	env.ESURL = os.Getenv("ELASTICSEARCH_URL")
	env.MasterUsername = os.Getenv("MASTER_USERNAME")
	env.MasterPassword = os.Getenv("MASTER_PASSWORD")
}

func HandleSearch(w http.ResponseWriter, r *http.Request) {
	// Get the search query from the URL
	query := r.URL.Query().Get("query")
	log.Println("Search query:", query)

	// Prepare the OpenSearch query
	esURL := env.ESURL + "/your-index-name/_search"
	queryJSON := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"description": query,
			},
		},
	}
	queryBytes, err := json.Marshal(queryJSON)
	if err != nil {
		log.Println("Error marshalling query:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Send the request to Elasticsearch using curl
	cmd := fmt.Sprintf(`curl -X POST -u %s:%s -H "Content-Type: application/json" -d '%s' %s`, env.MasterUsername, env.MasterPassword, string(queryBytes), esURL)
	resp, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		log.Println("Error executing curl command:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Write the response
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(resp); err != nil {
		log.Println("Error writing response:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
