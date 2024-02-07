package oauth

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"lorallabs.com/oauth-server/internal/config"
	"lorallabs.com/oauth-server/internal/store"
	"lorallabs.com/oauth-server/pkg/db"
)

type OAuthHandler struct {
	Config *config.Config
	Store  *store.Store
}

func NewOAuthHandler(cfg *config.Config, store *store.Store) *OAuthHandler {
	return &OAuthHandler{
		Config: cfg,
		Store:  store,
	}
}

func (h *OAuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	scope := "profile.compact cart.basic:write"
	authURL := "https://api.kroger.com/v1/connect/oauth2/authorize" +
		"?response_type=code" +
		"&client_id=" + h.Config.ClientID +
		"&redirect_uri=" + url.QueryEscape(h.Config.RedirectURI) +
		"&scope=" + url.QueryEscape(scope)
	http.Redirect(w, r, authURL, http.StatusFound)
}

func (h *OAuthHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}

	token, err := h.exchangeCodeForToken(code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Store the token in the database
	newToken := db.Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       time.Now().Add(time.Second * time.Duration(token.Expiry)).Unix(), // Unix time
	}
	if err := h.Store.SaveToken(&newToken); err != nil {
		http.Error(w, "Failed to save token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Authentication successful"))
}

// RefreshAccessToken uses the refresh token to obtain a new access token when the current one expires
func (h *OAuthHandler) RefreshAccessToken(refreshToken string) (*db.Token, error) {
	client := &http.Client{}
	data := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.kroger.com/v1/connect/oauth2/token", bytes.NewBuffer(dataBytes))
	if err != nil {
		return nil, err
	}

	authHeader := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", h.Config.ClientID, h.Config.ClientSecret))))
	req.Header.Add("Authorization", authHeader)
	req.Header.Add("Content-Type", "application/json")

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

	newToken := &db.Token{
		AccessToken:  tokenResponse.AccessToken,
		RefreshToken: tokenResponse.RefreshToken,
		Expiry:       time.Now().Add(time.Second * time.Duration(tokenResponse.ExpiresIn)).Unix(),
	}

	return newToken, nil
}

func (h *OAuthHandler) exchangeCodeForToken(code string) (*db.Token, error) {
	client := &http.Client{}
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", h.Config.RedirectURI)

	req, err := http.NewRequest("POST", "https://api.kroger.com/v1/connect/oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	// Add Authorization header
	auth := base64.StdEncoding.EncodeToString([]byte(h.Config.ClientID + ":" + h.Config.ClientSecret))
	req.Header.Add("Authorization", "Basic "+auth)
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
		Expiry       int64  `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return nil, err
	}

	return &db.Token{
		AccessToken:  tokenResponse.AccessToken,
		RefreshToken: tokenResponse.RefreshToken,
		Expiry:       tokenResponse.Expiry,
	}, nil
}

func (h *OAuthHandler) EnsureValidToken(tokenId uint) error {
	// Example logic to retrieve the current token from your store
	currentToken, err := h.Store.GetTokenByID(tokenId)
	if err != nil {
		return err
	}

	// Check if the current access token has expired
	if time.Now().Unix() > currentToken.Expiry {
		// Refresh the token
		newToken, err := h.RefreshAccessToken(currentToken.RefreshToken)
		if err != nil {
			return err
		}

		// Save the new token
		err = h.Store.SaveToken(newToken)
		if err != nil {
			return err
		}
	}

	return nil
}
