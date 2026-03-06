package models

import (
	"time"

	"github.com/google/uuid"
)

// APIKey représente une clé d'API pour l'authentification programmatique
type APIKey struct {
	ID        uuid.UUID  `json:"id" gorm:"primary_key;type:uuid;default:gen_random_uuid()"`
	Name      string     `json:"name" gorm:"size:100;not null"`
	KeyHash   string     `json:"-" gorm:"size:255;not null;uniqueIndex"`
	KeyPrefix string     `json:"key_prefix" gorm:"size:10;not null"`
	UserID    uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	User      User       `json:"-" gorm:"foreignKey:UserID"`
	ExpiresAt *time.Time `json:"expires_at"`

	CreatedAt time.Time `json:"created_at"`
}
