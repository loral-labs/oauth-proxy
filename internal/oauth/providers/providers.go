package providers

import "github.com/google/uuid"

type Provider interface {
	GetAuthURL(userId uuid.UUID, clientRedirectURI string) string
	ExchangeCodeForToken(code string) (*Token, error)
	RefreshToken(refreshToken string) (*Token, error)
}

type Token struct {
	AccessToken  string
	RefreshToken string
	Expiry       int64
}
