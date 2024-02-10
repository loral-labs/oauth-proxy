package oauthserver

import (
	"fmt"
	ory "github.com/ory/client-go"
	"os"
)

// token string | The string value of the token. For access tokens, this is the \\\"access_token\\\" value returned from the token endpoint defined in OAuth 2.0. For refresh tokens, this is the \\\"refresh_token\\\" value returned.
// scope string | An optional, space separated list of required scopes. If the access token was not granted one of the scopes, the result of active will be false. (optional)
func (o *OryClient) IntrospectToken(token string, scope string) *ory.IntrospectedOAuth2Token {

	resp, r, err := o.ory.OAuth2API.IntrospectOAuth2Token(o.ctx).Token(token).Scope(scope).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `OAuth2API.IntrospectOAuth2Token``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `IntrospectOAuth2Token`: IntrospectedOAuth2Token
	return resp
}
