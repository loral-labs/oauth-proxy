package oauthserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	ory "github.com/ory/client-go"
	"lorallabs.com/oauth-server/internal/store"
	"lorallabs.com/oauth-server/internal/types"
)

// token string | The string value of the token. For access tokens, this is the \\\"access_token\\\" value returned from the token endpoint defined in OAuth 2.0. For refresh tokens, this is the \\\"refresh_token\\\" value returned.
// scope string | An optional, space separated list of required scopes. If the access token was not granted one of the scopes, the result of active will be false. (optional)
func (o *OryClient) IntrospectToken(token string, scope string) *ory.IntrospectedOAuth2Token {

	resp, r, err := o.ory.OAuth2API.IntrospectOAuth2Token(o.ctx).Token(token).Scope(scope).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `OAuth2API.IntrospectOAuth2Token``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	return resp
}

func (o *OryClient) ListAppsHandler(w http.ResponseWriter, r *http.Request) {
	// get the bearer token from the request
	authHeader := r.Header.Get("Authorization")
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "Unauthorized - Invalid token format", http.StatusUnauthorized)
		return
	}
	token := parts[1]

	store := o.ctx.Value(types.StoreKey).(*store.Store)
	resp := *o.IntrospectToken(token, "")
	scope := resp.GetScope()
	scopes := strings.Split(scope, " ")

	sub := resp.GetSub()
	userID, err := uuid.Parse(sub)
	if err != nil {
		http.Error(w, "Invalid User ID", http.StatusBadRequest)
		return
	}

	providers := make(map[string]bool)
	for _, s := range scopes {
		if s == "openid" || s == "offline_access" {
			continue
		}
		// check if the user has authenticated to the provider
		exists, err := store.CheckValidProviderToken(userID, s)
		if err != nil {
			http.Error(w, "Error checking for valid provider token", http.StatusInternalServerError)
			return
		}
		providers[s] = exists
	}

	// Marshal the providers map into a JSON string
	providersJSON, err := json.Marshal(providers)
	if err != nil {
		http.Error(w, "Error marshaling providers", http.StatusInternalServerError)
		return
	}

	// Write the JSON response
	w.Header().Set("Content-Type", "application/json")
	w.Write(providersJSON)
}
