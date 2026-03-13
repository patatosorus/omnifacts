package models

import (
	"time"

	"github.com/google/uuid"
)

// PermissionLevel définit le niveau d'accès à un dépôt
type PermissionLevel string

const (
	PermRead  PermissionLevel = "read"
	PermWrite PermissionLevel = "write"
	PermAdmin PermissionLevel = "admin"
)

// Permission définit les droits d'un utilisateur sur un dépôt spécifique
type Permission struct {
	ID           uuid.UUID       `json:"id" gorm:"primary_key;type:uuid;default:gen_random_uuid()"`
	UserID       uuid.UUID       `json:"user_id" gorm:"type:uuid;not null;uniqueIndex:idx_user_repo"`
	User         User            `json:"-" gorm:"foreignKey:UserID"`
	RepositoryID uuid.UUID       `json:"repository_id" gorm:"type:uuid;not null;uniqueIndex:idx_user_repo"`
	Repository   Repository      `json:"-" gorm:"foreignKey:RepositoryID"`
	Level        PermissionLevel `json:"level" gorm:"size:20;not null"`

	CreatedAt time.Time `json:"created_at"`
}

// CanRead vérifie si la permission permet la lecture
func (p PermissionLevel) CanRead() bool {
	return p == PermRead || p == PermWrite || p == PermAdmin
}

// CanWrite vérifie si la permission permet l'écriture
func (p PermissionLevel) CanWrite() bool {
	return p == PermWrite || p == PermAdmin
}

// CanAdmin vérifie si la permission permet l'administration
func (p PermissionLevel) CanAdmin() bool {
	return p == PermAdmin
}
