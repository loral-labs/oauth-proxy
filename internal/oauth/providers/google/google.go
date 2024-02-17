package google

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"lorallabs.com/oauth-server/internal/oauth/providers"
)

type GoogleProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Scopes       string
}

type State struct {
	Nonce             string
	UserID            string
	ClientRedirectURI string
}

// Ensures clientRedirectURI and userID query params are set
func (k *GoogleProvider) URLParser(u *url.URL) {
	query := u.Query()
	state := query.Get("state")
	var stateObj State
	err := json.Unmarshal([]byte(state), &stateObj)
	if err != nil {
		log.Printf("Error unmarshalling state: %s", err)
	}
	query.Set("clientRedirectURI", stateObj.ClientRedirectURI)
	query.Set("userID", stateObj.UserID)
	u.RawQuery = query.Encode()
}

func (k *GoogleProvider) GetAuthURL(userID uuid.UUID, clientRedirectURI string) string {
	data := url.Values{}
	data.Set("redirect_uri", k.RedirectURI)
	data.Set("response_type", "code")
	data.Set("access_type", "offline")
	data.Set("client_id", k.ClientID)
	data.Set("scope", k.Scopes)
	state := State{
		Nonce:             uuid.New().String(),
		UserID:            userID.String(),
		ClientRedirectURI: clientRedirectURI,
	}
	stateString, err := json.Marshal(state)
	if err != nil {
		log.Printf("Error marshalling state: %s", err)
	}
	data.Set("state", string(stateString))

	authUrl := fmt.Sprintf("https://accounts.google.com/o/oauth2/v2/auth?%s", data.Encode())
	log.Printf("Auth URL: %s", authUrl)
	return authUrl
}

func (k *GoogleProvider) ExchangeCodeForToken(code string) (*providers.Token, error) {
	client := &http.Client{}
	data := url.Values{}
	data.Set("client_id", k.ClientID)
	data.Set("client_secret", k.ClientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", k.RedirectURI)

	req, err := http.NewRequest("POST", "https://oauth2.googleapis.com/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	log.Printf("Body: %s\n%s", code, body)

	var tokenResponse struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return nil, err
	}

	return &providers.Token{
		AccessToken:  tokenResponse.AccessToken,
		RefreshToken: tokenResponse.RefreshToken,
		Expiry:       tokenResponse.ExpiresIn,
	}, nil
}

func (k *GoogleProvider) RefreshToken(refreshToken string) (*providers.Token, error) {
	client := &http.Client{}
	data := url.Values{}
	data.Set("client_id", k.ClientID)
	data.Set("client_secret", k.ClientSecret)
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	req, err := http.NewRequest("POST", "https://api.kroger.com/v1/connect/oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error refreshing token: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tokenResponse struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return nil, err
	}

	return &providers.Token{
		AccessToken:  tokenResponse.AccessToken,
		RefreshToken: tokenResponse.RefreshToken,
		Expiry:       tokenResponse.ExpiresIn,
	}, nil
}
