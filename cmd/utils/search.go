package utils

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

type SearchResponse struct {
	Hits Hits `json:"hits"`
}

type Hits struct {
	Hits []Hit `json:"hits"`
}

type Hit struct {
	Source Source `json:"_source"`
}

type Source struct {
	FunctionStr string `json:"function_str"`
}

func processSearchHits(searchHits []Hit) (interface{}, error) {
	response := make(map[string]interface{})
	response["items"] = make([]string, 0)

	for _, hit := range searchHits {
		hitSource := hit.Source
		response["items"] = append(response["items"].([]string), hitSource.FunctionStr)
	}

	return response, nil
}

func HandleSearch(w http.ResponseWriter, r *http.Request) {
	// Get the search query from the URL
	query := r.URL.Query().Get("query")
	domain := r.URL.Query().Get("domain")

	log.Default().Printf("Search hit for query: %s", query)

	// Prepare the OpenSearch query
	esURL := env.ESURL + "/function-index/_search"
	queryJSON := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"match": map[string]string{
							"function_str": query,
						},
					},
				},
				"filter": map[string]interface{}{
					"term": map[string]string{
						"domain": domain,
					},
				},
			},
		},
		"_source": []string{"function_str"},
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
