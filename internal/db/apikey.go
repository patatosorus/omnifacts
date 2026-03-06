package db

import (
	"omnifacts/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type APIKeyDB interface {
	Create(key *models.APIKey) error
	FindByHash(keyHash string) (*models.APIKey, error)
	ListByUser(userID uuid.UUID) ([]models.APIKey, error)
	Delete(id uuid.UUID) error
}

type apiKeyDB struct {
	db *gorm.DB
}

func NewAPIKeyDB(db *gorm.DB) APIKeyDB {
	return &apiKeyDB{db: db}
}

func (r *apiKeyDB) Create(key *models.APIKey) error {
	return r.db.Create(key).Error
}

func (r *apiKeyDB) FindByHash(keyHash string) (*models.APIKey, error) {
	var key models.APIKey
	err := r.db.Preload("User").First(&key, "key_hash = ?", keyHash).Error
	return &key, err
}

func (r *apiKeyDB) ListByUser(userID uuid.UUID) ([]models.APIKey, error) {
	var keys []models.APIKey
	err := r.db.Where("user_id = ?", userID).Find(&keys).Error
	return keys, err
}

func (r *apiKeyDB) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.APIKey{}, id).Error
}
