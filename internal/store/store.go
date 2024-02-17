package store

import (
	"log"

	"github.com/google/uuid"
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
		&schema.Provider{},
		&schema.ProviderToken{},
		&schema.APIKey{},
		&schema.Client{},
		&schema.ClientGrants{},
	)
	if err != nil {
		return nil, err
	}

	return &Store{DB: db}, nil
}

func (s *Store) CheckValidProviderToken(userId uuid.UUID, providerName string) (bool, error) {
	var token schema.ProviderToken

	err := s.DB.Joins("JOIN providers ON providers.id = provider_tokens.provider_id").
		Where("provider_tokens.user_id = ?", userId.String()).
		Where("providers.name = ?", providerName).
		First(&token).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		// Handle other errors
		log.Fatalf("Error checking for valid provider token: %v\n", err)
		return false, err
	}

	return true, nil
}
