package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		// handle error
	}

	ctx := context.Background()

	// Step 1: Create a new HTTP client
	httpClient, err := httpclient.NewHTTPClient()
	if err != nil {
		// handle error
	}

	// Step 2: Obtain an OAuth2 token
	_, err = getOAuth2Token(ctx, *httpClient)
	if err != nil {
		// handle error
	}
	// fmt.Println(token.AccessToken)
	// // Step 3: Configure the HTTP client with the OAuth2 token
	// authClient, err := httpclient.NewHTTPClient(
	// 	httpclient.WithAuthToken(token.AccessToken),
	// 	// ... other configurations
	// )
	// if err != nil {
	// 	// handle error
	// }

	// Use authClient for making authenticated requests to Kroger API
}

// OAuth2TokenResponse represents the JSON response from OAuth token endpoint
type OAuth2TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// getOAuth2Token retrieves an OAuth2 token using client credentials
func getOAuth2Token(ctx context.Context, client http.Client) (*OAuth2TokenResponse, error) {
	queryParams := url.Values{}
	queryParams.Set("scope", os.Getenv("SCOPES"))
	queryParams.Set("client_id", os.Getenv("CLIENT_ID"))
	queryParams.Set("redirect_uri", os.Getenv("REDIRECT_URI"))
	queryParams.Set("response_type", "code")

	krogerTokenURLWithParams := fmt.Sprintf("%s?%s", os.Getenv("ISSUER"), queryParams.Encode())

	// print the URL
	fmt.Println("URL: ", krogerTokenURLWithParams)

	req, err := http.NewRequest("GET", krogerTokenURLWithParams, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(os.Getenv("CLIENT_ID"), os.Getenv("CLIENT_SECRET"))

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error: ", resp.Status)
		// read and print the response body
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		fmt.Println(string(respBody))
	}

	// var redirectURL string
	// print the content-type of the response
	contentType := resp.Header.Get("Content-Type")
	fmt.Println("Content Type: ", contentType)

	return nil, nil
}
