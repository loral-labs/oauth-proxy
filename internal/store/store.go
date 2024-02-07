package store

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	schema "lorallabs.com/oauth-server/pkg/db" // Make sure this import path matches your module structure
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

	// AutoMigrate your schema here
	err = db.AutoMigrate(&schema.Token{})
	if err != nil {
		return nil, err
	}

	return &Store{DB: db}, nil
}

// SaveToken saves a new OAuth token to the database
func (s *Store) SaveToken(token *schema.Token) error {
	return s.DB.Create(token).Error
}

// GetTokenByID retrieves a token by its ID from the database.
func (s *Store) GetTokenByID(id uint) (*schema.Token, error) {
	var token schema.Token
	err := s.DB.First(&token, id).Error
	if err != nil {
		return nil, err // This will include gorm.ErrRecordNotFound if no token is found
	}
	return &token, nil
}
