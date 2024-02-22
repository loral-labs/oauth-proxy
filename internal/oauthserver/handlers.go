package oauthserver

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func (o *OryClient) RegisterOAuthServerHandlers(handler *mux.Router) {
	handler.HandleFunc("/client/create", func(w http.ResponseWriter, r *http.Request) {
		type CreateClientRequest struct {
			Name         string   `json:"name"`
			RedirectUris []string `json:"redirect_uris"`
			Scopes       []string `json:"scopes"`
		}
		var request CreateClientRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		clientID, clientSecret := o.CreateClient(request.Name, request.RedirectUris, request.Scopes)
		client := struct {
			ID     string `json:"id"`
			Secret string `json:"secret"`
		}{
			ID:     clientID,
			Secret: clientSecret,
		}
		clientJSON, err := json.Marshal(client)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(clientJSON)
	}).Methods("POST")

	handler.HandleFunc("/client/edit/name", func(w http.ResponseWriter, r *http.Request) {
		// get request
		type ReplaceNameRequest struct {
			ID     string `json:"id"`
			Secret string `json:"secret"`
			Name   string `json:"name"`
		}
		var request ReplaceNameRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		o.ReplaceName(request.ID, request.Secret, request.Name)
	})

	handler.HandleFunc("/client/edit/scope", func(w http.ResponseWriter, r *http.Request) {
		// get request
		type AddScopeRequest struct {
			ID     string `json:"id"`
			Secret string `json:"secret"`
			Scope  string `json:"name"`
			Add    bool   `json:"add"`
		}
		var request AddScopeRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if request.Add {
			o.AddScope(request.ID, request.Secret, request.Scope)
		} else {
			o.RemoveScope(request.ID, request.Secret, request.Scope)
		}
	})

	handler.HandleFunc("/client/edit/redirectUris", func(w http.ResponseWriter, r *http.Request) {
		// get request
		type ReplaceRedirectUrisRequest struct {
			ID     string   `json:"id"`
			Secret string   `json:"secret"`
			Uris   []string `json:"uris"`
		}
		var request ReplaceRedirectUrisRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		o.ReplaceRedirectUris(request.ID, request.Secret, request.Uris)
	})
}
