package store

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	schema "lorallabs.com/oauth-server/pkg/db"
)

// Store encapsulates DB operations
type Store struct {
	DB *gorm.DB
}

// NewStore creates a new instance of Store with a database connection
func NewStore(connectionString string) (*Store, error) {
	db, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Executing the raw SQL to create the UUID extension
	err = db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";").Error
	if err != nil {
		log.Fatalf("Failed to create UUID extension: %v", err)
	}

	// AutoMigrate your schema here
	err = db.AutoMigrate(
		&schema.User{},
		&schema.APIKey{},
		&schema.Client{},
		&schema.ClientGrants{},
		&schema.Provider{},
		&schema.ProviderToken{},
	)
	if err != nil {
		return nil, err
	}

	return &Store{DB: db}, nil
}

// SaveToken saves a new OAuth token to the database
func (s *Store) SaveToken(token *schema.ProviderToken) error {
	return s.DB.Create(token).Error
}

// GetTokenByID retrieves a token by a client-provided from the database.
func (s *Store) GetTokenByClientID(id string, provider string) (*schema.ProviderToken, error) {
	var clientRecord *schema.Client
	err := s.DB.Where("identifier = ?", id).First(&clientRecord).Error
	if err != nil {
		return nil, err
	}
	var token schema.ProviderToken
	err = s.DB.Where("user_id = ? AND provider = ?", clientRecord.UserID, provider).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}
