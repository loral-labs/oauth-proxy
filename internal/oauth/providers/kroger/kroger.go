package kroger

import (
	"encoding/base64"
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

type KrogerProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Scopes       string
}

func NewKrogerProvider(clientID, clientSecret, redirectURI string, scopes string) *KrogerProvider {
	return &KrogerProvider{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURI:  redirectURI,
		Scopes:       scopes,
	}
}

func (k *KrogerProvider) GetName() string {
	return "kroger"
}

func (k *KrogerProvider) GetAuthURL(userID uuid.UUID) string {
	scope := k.Scopes
	k.RedirectURI = fmt.Sprintf("%s?userID=%s", k.RedirectURI, userID.String())
	authUrl := fmt.Sprintf("https://api.kroger.com/v1/connect/oauth2/authorize?response_type=code&client_id=%s&redirect_uri=%s&scope=%s",
		k.ClientID, url.QueryEscape(k.RedirectURI), url.QueryEscape(scope))
	log.Printf("Auth URL: %s", authUrl)
	return authUrl
}

func (k *KrogerProvider) ExchangeCodeForToken(code string) (*providers.Token, error) {
	client := &http.Client{}
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", k.RedirectURI)

	req, err := http.NewRequest("POST", "https://api.kroger.com/v1/connect/oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	authHeader := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", k.ClientID, k.ClientSecret))))
	req.Header.Add("Authorization", authHeader)
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

func (k *KrogerProvider) RefreshToken(refreshToken string) (*providers.Token, error) {
	client := &http.Client{}
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	req, err := http.NewRequest("POST", "https://api.kroger.com/v1/connect/oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	authHeader := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", k.ClientID, k.ClientSecret))))
	req.Header.Add("Authorization", authHeader)
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
