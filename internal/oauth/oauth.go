package oauth

import (
	"net/http"
	"time"

	"lorallabs.com/oauth-server/internal/config"
	"lorallabs.com/oauth-server/internal/oauth/providers"
	"lorallabs.com/oauth-server/internal/oauth/providers/kroger"
	"lorallabs.com/oauth-server/internal/store"
	schema "lorallabs.com/oauth-server/pkg/db"
)

type OAuthHandler struct {
	ProviderMap map[string]providers.Provider
	Store       *store.Store
}

func NewOAuthHandler(config *config.Config, store *store.Store) *OAuthHandler {
	providerMap := InitializeProviders(config)
	return &OAuthHandler{
		ProviderMap: providerMap,
		Store:       store,
	}
}

// InitializeProviders sets up OAuth providers
func InitializeProviders(config *config.Config) map[string]providers.Provider {
	providers := map[string]providers.Provider{
		"kroger": &kroger.KrogerProvider{ClientID: config.KrogerClientID, ClientSecret: config.KrogerClientSecret, RedirectURI: config.KrogerRedirectURI, Scopes: config.KrogerScopes},
		// Initialize other providers similarly
	}
	return providers
}

// HandleAuth initiates the OAuth flow for a given provider
func (h *OAuthHandler) HandleAuth(providerName string, w http.ResponseWriter, r *http.Request) {
	provider, exists := h.ProviderMap[providerName]
	if !exists {
		http.Error(w, "Unsupported provider", http.StatusBadRequest)
		return
	}
	url := provider.GetAuthURL()
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// HandleCallback handles the callback for a given provider
func (h *OAuthHandler) HandleCallback(providerName string, w http.ResponseWriter, r *http.Request) {
	provider, exists := h.ProviderMap[providerName]
	if !exists {
		http.Error(w, "Unsupported provider", http.StatusBadRequest)
		return
	}
	code := r.URL.Query().Get("code")
	token, err := provider.ExchangeCodeForToken(code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// search for the provider in the database
	dbProvider := &schema.Provider{}
	err = h.Store.DB.Where("name = ?", providerName).First(&dbProvider).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a ProviderToken from the token and store it
	providerToken := &schema.ProviderToken{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       time.Now().Add(time.Duration(token.Expiry) * time.Second).Unix(),
		UserID:       1, // Replace with the actual user ID
		ProviderID:   dbProvider.ID,
	}

	// If a token already exists for the user and provider, update it
	var existingToken schema.ProviderToken
	err = h.Store.DB.Where("user_id = ? AND provider_id = ?", providerToken.UserID, providerToken.ProviderID).First(&existingToken).Error
	if err == nil {
		providerToken.ID = existingToken.ID
		providerToken.UUID = existingToken.UUID
		err = h.Store.DB.Save(providerToken).Error
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	err = h.Store.DB.Create(providerToken).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *OAuthHandler) HandleGetToken(providerName string, userID string) string {
	// Find the provider in the database
	dbProvider := &schema.Provider{}
	err := h.Store.DB.Where("name = ?", providerName).First(&dbProvider).Error
	if err != nil {
		return err.Error()
	}

	// Find the token for the user and provider
	var providerToken schema.ProviderToken
	err = h.Store.DB.Where("user_id = ? AND provider_id = ?", userID, dbProvider.ID).First(&providerToken).Error
	if err != nil {
		return err.Error()
	}

	// check if the token is expired
	if time.Now().Unix() > providerToken.Expiry {
		// refresh the token
		provider, exists := h.ProviderMap[providerName]
		if !exists {
			return err.Error()
		}
		token, err := provider.RefreshToken(providerToken.RefreshToken)
		if err != nil {
			return err.Error()
		}
		// update the token in the database
		providerToken.AccessToken = token.AccessToken
		providerToken.RefreshToken = token.RefreshToken
		providerToken.Expiry = time.Now().Add(time.Duration(token.Expiry) * time.Second).Unix()
		err = h.Store.DB.Save(&providerToken).Error
		if err != nil {
			return err.Error()
		}
	}

	return providerToken.AccessToken
}
