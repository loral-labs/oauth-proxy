package db

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a system user with relationships to ProviderTokens and ClientTokens
type User struct {
	ID             uint      `gorm:"primaryKey"`
	UUID           uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	Username       string    `gorm:"uniqueIndex"`
	Email          string    `gorm:"uniqueIndex"`
	Password       string    // It's recommended to store a hashed password
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
	ProviderTokens []ProviderToken
}

// ProviderToken represents OAuth tokens provided by external providers, related to a User and an Application
type ProviderToken struct {
	ID            uint      `gorm:"primaryKey"`
	UUID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	AccessToken   string
	RefreshToken  string
	Expiry        int64 // Unix time
	UserID        uint  // Foreign key for User
	ApplicationID uint  // Foreign key for Application
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

// Application represents an external application that has ProviderTokens
type Provider struct {
	ID             uint   `gorm:"primaryKey"`
	Name           string `gorm:"unique"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt  `gorm:"index"`
	ProviderTokens []ProviderToken // Relation back to ProviderTokens
}
