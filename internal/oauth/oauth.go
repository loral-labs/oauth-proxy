package oauth

import (
	"log"
	"net/http"

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
		"kroger": &kroger.KrogerProvider{ClientID: config.KrogerClientID, ClientSecret: config.KrogerClientSecret, RedirectURI: config.KrogerRedirectURI},
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

	// Process token (store it, use it, etc.)
	log.Default().Printf("%s Token: %+v", providerName, token)

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
		Expiry:       token.Expiry,
		UserID:       1, // Replace with the actual user ID
		ProviderID:   dbProvider.ID,
	}

	err = h.Store.DB.Create(providerToken).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
