package models

import (
	"time"

	"github.com/google/uuid"
)

// Role définit le rôle global d'un utilisateur
type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

// User représente un utilisateur du registre
type User struct {
	ID           uuid.UUID `json:"id" gorm:"primary_key;type:uuid;default:gen_random_uuid()"`
	Username     string    `json:"username" gorm:"size:100;not null;uniqueIndex"`
	Email        string    `json:"email" gorm:"size:255;not null;uniqueIndex"`
	PasswordHash string    `json:"-" gorm:"size:255;not null"`
	Role         Role      `json:"role" gorm:"size:20;not null;default:'user'"`
	Active       bool      `json:"active" gorm:"default:true"`

	APIKeys []APIKey `json:"api_keys,omitempty" gorm:"foreignKey:UserID"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
