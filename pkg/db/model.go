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
	Clients        []Client
}

type Client struct {
	ID         uint      `gorm:"primaryKey"`
	UUID       uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	Name       string    `gorm:"unique"`
	Identifier string    `gorm:"unique"` // a client-provided unique identifier
	UserID     uint      // Foreign key for User
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

type Provider struct {
	ID             uint      `gorm:"primaryKey"`
	UUID           uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	Name           string    `gorm:"unique"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt  `gorm:"index"`
	ProviderTokens []ProviderToken `gorm:"foreignKey:ProviderID"` // Explicitly define the foreign key relationship
}
type ProviderToken struct {
	ID           uint      `gorm:"primaryKey"`
	UUID         uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	AccessToken  string
	RefreshToken string
	Expiry       int64 // Unix time
	UserID       uint  // Foreign key for User
	ProviderID   uint  // Foreign key for Provider, assuming this is the missing link
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}
