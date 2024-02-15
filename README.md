## Registering a new provider

- {provider}/auth
- {provider}/auth/callback

## Development

- run `go run cmd/main.go` to migrate the db
- run `go run scripts/add_provider.go` to initialize a provider`
- run `go run cmd/main.go --lax_auth` to start the server with lax auth

- Running with --lax_auth flag to accept expired or out-of-scope tokens. A structurally correct token is still required to parse the user's identity.
