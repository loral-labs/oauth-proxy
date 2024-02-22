package oauthserver

import (
	"context"
	"fmt"
	"net/http"
	"os"

	ory "github.com/ory/client-go"
	"lorallabs.com/oauth-server/internal/types"
)

type OryClient struct {
	ory *ory.APIClient
	ctx context.Context
}

func NewOryClient(ctx context.Context) *OryClient {
	configuration := ory.NewConfiguration()
	configuration.Servers = []ory.ServerConfiguration{
		{
			URL: "https://fervent-cori-shm6sflkse.projects.oryapis.com", // Public API URL
		},
	}
	ory := ory.NewAPIClient(configuration)
	return &OryClient{ory: ory, ctx: ctx}
}

func (o *OryClient) CreateClient(clientName string, redirectUris []string) {
	oryAuthedContext := o.ctx
	oAuth2Client := *ory.NewOAuth2Client() // OAuth2Client |
	oAuth2Client.SetClientName(clientName)
	oAuth2Client.SetScope("openid offline")
	oAuth2Client.SetSkipConsent(true)
	oAuth2Client.SetRedirectUris(redirectUris)
	oAuth2Client.SetGrantTypes([]string{"authorization_code", "refresh_token"})
	oAuth2Client.SetResponseTypes([]string{"code", "token"})

	resp, r, err := o.ory.OAuth2API.CreateOAuth2Client(oryAuthedContext).OAuth2Client(oAuth2Client).Execute()
	if err != nil {
		switch r.StatusCode {
		case http.StatusConflict:
			fmt.Fprintf(os.Stderr, "Conflict when creating oAuth2Client: %v\n", err)
		default:
			fmt.Fprintf(os.Stderr, "Error when calling `OAuth2Api.CreateOAuth2Client`: %v\n", err)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		}
	}
	// response from `CreateOAuth2Client`: OAuth2Client
	fmt.Fprintf(os.Stdout, "Created client with name %s\n", resp.GetClientName())
}

func (o *OryClient) ListClients(clientName string) {
	oryAuthedContext := o.ctx

	clients, r, err := o.ory.OAuth2API.ListOAuth2Clients(oryAuthedContext).ClientName(clientName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AdminApi.ListOAuth2Clients``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	fmt.Fprintf(os.Stdout, "We have %d clients\n", len(clients))
	for i, client := range clients {
		fmt.Fprintf(os.Stdout, "Client %d id: %s, name: %s\n", i+1, *client.ClientId, client.GetClientName())
		// Add any additional information you want to display about the client
	}
}

func (o *OryClient) PatchClient(id string, clientSecret string, op types.Operation, path string, value string) {
	oryAuthedContext := o.ctx

	jsonPatch := []ory.JsonPatch{*ory.NewJsonPatch(string(op), path)} // []JsonPatch | OAuth 2.0 Client JSON Patch Body
	jsonPatch[0].SetValue(value)

	// verify the client secret
	client := o.GetClient(id)
	if client.GetClientSecret() != clientSecret {
		fmt.Fprintf(os.Stderr, "Client secret does not match\n")
		return
	}

	resp, r, err := o.ory.OAuth2API.PatchOAuth2Client(oryAuthedContext, id).JsonPatch(jsonPatch).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AdminApi.UpdateOAuth2Client``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	fmt.Fprintf(os.Stdout, "Updated client with name %s\n", resp.GetClientName())
}

func (o *OryClient) GetClient(id string) *ory.OAuth2Client {
	resp, r, err := o.ory.OAuth2API.GetOAuth2Client(o.ctx, id).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `OAuth2API.GetOAuth2Client``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	return resp
}
