package oauthserver

import (
	"strings"
)

func (o *OryClient) AddScope(id string, clientSecret string, scope string) {
	client := o.GetClient(id)
	scope = *client.Scope + " " + scope
	o.PatchClient(id, clientSecret, "replace", "/scope", scope)
}

func (o *OryClient) RemoveScope(id string, clientSecret string, scope string) {
	client := o.GetClient(id)
	scope = strings.Replace(*client.Scope, scope, "", -1)
	o.PatchClient(id, clientSecret, "replace", "/scope", scope)
}

func (o *OryClient) ReplaceName(id string, clientSecret string, name string) {
	o.PatchClient(id, clientSecret, "replace", "/client_name", name)
}

func (o *OryClient) ReplaceRedirectUris(id string, clientSecret string, redirectUris []string) {
	o.PatchClient(id, clientSecret, "replace", "/redirect_uris", redirectUris)
}
