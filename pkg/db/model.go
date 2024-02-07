package db

// Token represents the OAuth tokens we need to store
type Token struct {
	ID           uint `gorm:"primaryKey"`
	AccessToken  string
	RefreshToken string
	Expiry       int64 // Unix time
}
