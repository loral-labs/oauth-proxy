package oauthserver

import (
	"context"
	"fmt"
	"log"
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

func (o *OryClient) Authorize() {
	oryAuthedContext := o.ctx
	apiClient := o.ory

	resp, r, err := apiClient.OAuth2API.OAuth2Authorize(oryAuthedContext).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `OAuth2API.OAuth2Authorize``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `OAuth2Authorize`: ErrorOAuth2
	fmt.Fprintf(os.Stdout, "Response from `OAuth2API.OAuth2Authorize`: %v\n", resp)
}

func (o *OryClient) TokenServer() {
	oryAuthedContext := o.ctx
	apiClient := o.ory

	grantType := "refresh_token"                                                                                     // string | refresh_token
	clientId := "aca314fc-8db0-4840-857c-99343e7d40c7"                                                               // string |  (optional)
	redirectUri := "http://127.0.0.1:4446/callback"                                                                  // string |  (optional)
	refreshToken := "ory_rt_VJUb-5vv3rg4vKmY6Qr2vNK_AD_n1sVqh6n7jSAVCoo.at88DWAluCmCx2mPdGFIgQ7MhQley8fzu0baBAJUGLg" // string |  (optional)

	resp, r, err := apiClient.OAuth2API.Oauth2TokenExchange(oryAuthedContext).GrantType(grantType).ClientId(clientId).RedirectUri(redirectUri).RefreshToken(refreshToken).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `OAuth2API.Oauth2TokenExchange``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `Oauth2TokenExchange`: OAuth2TokenExchange
	fmt.Fprintf(os.Stdout, "Response from `OAuth2API.Oauth2TokenExchange`: %v\n", resp)
}

func (o *OryClient) CreateClient() {
	oryAuthedContext := o.ctx
	clientName := "example_client"
	oAuth2Client := *ory.NewOAuth2Client() // OAuth2Client |
	oAuth2Client.SetClientName(clientName)

	log.Default().Println("Ory API Key: ", os.Getenv("ORY_API_KEY"))

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

func (o *OryClient) PatchClient(id string, op types.Operation, path string, value string) {
	oryAuthedContext := o.ctx

	jsonPatch := []ory.JsonPatch{*ory.NewJsonPatch(string(op), path)} // []JsonPatch | OAuth 2.0 Client JSON Patch Body
	jsonPatch[0].SetValue(value)

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
