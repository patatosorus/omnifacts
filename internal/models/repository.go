package models

import (
	"time"

	"github.com/google/uuid"
)

// StorageMode définit le mode de stockage d'un dépôt
type StorageMode string

const (
	StorageLocal  StorageMode = "local"  // Stockage dans le registre OCI local (zot)
	StorageMirror StorageMode = "mirror" // Proxy cache vers un registre distant (pull-through)
	StorageRemote StorageMode = "remote" // Proxy pur vers un registre distant
)

// Repository représente un dépôt logique contenant des artefacts d'un seul type
type Repository struct {
	ID           uuid.UUID   `json:"id" gorm:"primary_key;type:uuid;default:gen_random_uuid()"`
	Name         string      `json:"name" gorm:"size:255;not null;uniqueIndex"`
	Description  string      `json:"description" gorm:"size:1000"`
	ArtefactType string      `json:"artefact_type" gorm:"size:50;not null"` // docker, helm, pypi, npm, etc.
	StorageMode  StorageMode `json:"storage_mode" gorm:"size:20;not null;default:'local'"`

	// Configuration pour mirror/remote
	UpstreamURL string `json:"upstream_url,omitempty" gorm:"size:500"` // URL du registre upstream (ex: https://registry-1.docker.io)

	Artefacts []Artefact `json:"artefacts,omitempty" gorm:"foreignKey:RepositoryID"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
