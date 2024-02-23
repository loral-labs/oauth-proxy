package oauthserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
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

type JWKS struct {
	Keys []JWK `json:"keys"`
}

type JWK struct {
	Kty string `json:"kty"`
	E   string `json:"e"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	N   string `json:"n"`
}

func (o *OryClient) CreateClient(clientName string, redirectUris []string, providerScopes []string) (clientID string, clientSecret string) {
	oryAuthedContext := o.ctx
	oAuth2Client := *ory.NewOAuth2Client()
	oAuth2Client.SetClientName(clientName)
	providerScopes = append(providerScopes, "openid", "offline_access")
	oAuth2Client.SetScope(strings.Join(providerScopes, " "))
	oAuth2Client.SetRedirectUris(redirectUris)
	oAuth2Client.SetGrantTypes([]string{"authorization_code", "refresh_token"})
	oAuth2Client.SetResponseTypes([]string{"code", "token"})
	secret := uuid.New().String()
	oAuth2Client.SetClientSecret(secret)
	// fix me, hacky to use this for auth
	jwk := JWK{
		Kty: "RSA",
		E:   "AQAB",
		Use: "sig",
		Kid: "ory-example",
		Alg: "RS256",
		N:   secret,
	}
	oAuth2Client.SetJwks(map[string]interface{}{"keys": []JWK{jwk}})

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
	return resp.GetClientId(), resp.GetClientSecret()
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

func (o *OryClient) PatchClient(id string, clientSecret string, op types.Operation, path string, value interface{}) error {
	oryAuthedContext := o.ctx

	jsonPatch := []ory.JsonPatch{*ory.NewJsonPatch(string(op), path)} // []JsonPatch | OAuth 2.0 Client JSON Patch Body
	jsonPatch[0].SetValue(value)

	// verify the client secret
	client := o.GetClient(id)
	jwksMap := client.GetJwks()
	jwksBytes, err := json.Marshal(jwksMap)
	if err != nil {
		return err
	}

	var jwks JWKS
	err = json.Unmarshal(jwksBytes, &jwks)
	if err != nil {
		return err
	}

	trueClientSecret := jwks.Keys[0].N
	if trueClientSecret != clientSecret {
		log.Printf("Client secret does not match: %s, %s", trueClientSecret, clientSecret)
		return fmt.Errorf("client secret does not match")
	}

	resp, r, err := o.ory.OAuth2API.PatchOAuth2Client(oryAuthedContext, id).JsonPatch(jsonPatch).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AdminApi.UpdateOAuth2Client``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	fmt.Fprintf(os.Stdout, "Updated client with name %s\n", resp.GetClientName())
	return nil
}

func (o *OryClient) GetClient(id string) *ory.OAuth2Client {
	resp, r, err := o.ory.OAuth2API.GetOAuth2Client(o.ctx, id).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `OAuth2API.GetOAuth2Client``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	return resp
}
