package oauth

import (
	"log"
	"net/http"

	"lorallabs.com/oauth-server/internal/config"
	"lorallabs.com/oauth-server/internal/oauth/providers"
	"lorallabs.com/oauth-server/internal/oauth/providers/kroger"
)

// InitializeProviders sets up OAuth providers
func InitializeProviders(config *config.Config) map[string]providers.Provider {
	providers := map[string]providers.Provider{
		"kroger": &kroger.KrogerProvider{ClientID: config.KrogerClientID, ClientSecret: config.KrogerClientSecret, RedirectURI: config.KrogerRedirectURI},
		// Initialize other providers similarly
	}
	return providers
}

// HandleAuth initiates the OAuth flow for a given provider
func HandleAuth(providerMap map[string]providers.Provider, providerName string, w http.ResponseWriter, r *http.Request) {
	provider, exists := providerMap[providerName]
	if !exists {
		http.Error(w, "Unsupported provider", http.StatusBadRequest)
		return
	}
	url := provider.GetAuthURL()
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// HandleCallback handles the callback for a given provider
func HandleCallback(providerMap map[string]providers.Provider, providerName string, w http.ResponseWriter, r *http.Request) {
	provider, exists := providerMap[providerName]
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
}
