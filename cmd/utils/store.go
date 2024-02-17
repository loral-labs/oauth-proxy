package utils

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

func HandleStore(w http.ResponseWriter, r *http.Request) {
	// Get the search query from the URL
	// Read the request body
	var requestBody map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		log.Println("Error decoding request body:", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	log.Default().Printf("Storing function for domain: %s", requestBody["domain"])

	// Prepare the OpenSearch query
	esURL := env.ESURL + "/function-index/_doc?pretty=true"

	docBytes, err := json.Marshal(requestBody)
	if err != nil {
		log.Println("Error marshalling query:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Send the request to Elasticsearch
	req, err := http.NewRequest("POST", esURL, bytes.NewBuffer(docBytes))
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

	response := map[string]bool{
		"success": true,
	}

	// Write the response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error encoding response:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
