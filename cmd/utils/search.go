package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"io"

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
	env.ESURL = os.Getenv("LORAL_ES_DOMAIN")
	env.MasterUsername = os.Getenv("LORAL_ES_DOMAIN_USER")
	env.MasterPassword = os.Getenv("LORAL_ES_DOMAIN_PSWD")
}

type SearchResponse struct {
	Hits Hits `json:"hits"`
}

type Hits struct {
	Hits []Hit `json:"hits"`
}

type Hit struct {
	Score  float64 `json:"_score"`
	Source Source  `json:"_source"`
}

type Source struct {
	ApiPath      string `json:"api_path"`
	ApiMethod    string `json:"api_method"`
	SpecFilePath string `json:"spec_file_path"`
}

func processSearchHits(searchHits []Hit) (interface{}, error) {
	response := make(map[string]interface{})

	for _, hit := range searchHits {
		if hit.Score < 0.75 {
			continue
		}
		httpSpec := hit.Source

		specPath := filepath.Join("data/kroger", httpSpec.SpecFilePath)
		spec, err := loadJSON(specPath)
		if err != nil {
			return nil, err
		}

		specMap, ok := spec.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("error: spec is not a map")
		}

		delete(specMap, "x-tagGroups")
		delete(specMap, "paths")
		specMap["components"] = make(map[string]interface{})
		specMap["endpoints"] = make(map[string]interface{})

		if _, ok := response[httpSpec.SpecFilePath]; !ok {
			response[httpSpec.SpecFilePath] = specMap
		}

		fullSpec, err := loadJSON(specPath)
		if err != nil {
			return nil, err
		}

		if _, ok := response[httpSpec.SpecFilePath].(map[string]interface{})["endpoints"].(map[string]interface{})[httpSpec.ApiPath]; !ok {
			response[httpSpec.SpecFilePath].(map[string]interface{})["endpoints"].(map[string]interface{})[httpSpec.ApiPath] = make(map[string]interface{})
		}

		endpointObject := fullSpec.(map[string]interface{})["paths"].(map[string]interface{})[httpSpec.ApiPath].(map[string]interface{})[httpSpec.ApiMethod]
		delete(endpointObject.(map[string]interface{}), "x-code-samples")
		response[httpSpec.SpecFilePath].(map[string]interface{})["endpoints"].(map[string]interface{})[httpSpec.ApiPath].(map[string]interface{})[httpSpec.ApiMethod] = endpointObject
	}

	return response, nil
}

func loadJSON(filePath string) (interface{}, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, file)
	if err != nil {
		return nil, err
	}
	data := buf.Bytes()
	if err != nil {
		return nil, err
	}

	var jsonData interface{}
	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

func HandleSearch(w http.ResponseWriter, r *http.Request) {
	// Get the search query from the URL
	query := r.URL.Query().Get("query")

	// Prepare the OpenSearch query
	esURL := env.ESURL + "/loral-http-index/_search"
	queryJSON := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"description": query,
			},
		},
		"_source": []string{"api_path", "api_method", "spec_file_path"},
	}
	queryBytes, err := json.Marshal(queryJSON)
	if err != nil {
		log.Println("Error marshalling query:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Send the request to Elasticsearch
	req, err := http.NewRequest("GET", esURL, bytes.NewBuffer(queryBytes))
	if err != nil {
		log.Println("Error creating request:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	req.SetBasicAuth(env.MasterUsername, env.MasterPassword)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending request to Elasticsearch:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read the response body
	var result SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Println("Error decoding response body:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	hits := result.Hits

	response, err := processSearchHits(hits.Hits)
	if err != nil {
		log.Println("Error processing search hits:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Write the response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error encoding response:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
