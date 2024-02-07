package providers

type Provider interface {
	GetAuthURL() string
	ExchangeCodeForToken(code string) (*Token, error)
	RefreshToken(refreshToken string) (*Token, error)
}

type Token struct {
	AccessToken  string
	RefreshToken string
	Expiry       int64
}
