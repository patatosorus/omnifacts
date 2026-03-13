package models

import (
	"time"

	"github.com/google/uuid"
)

// Artefact représente un artefact stocké dans le registre OCI
type Artefact struct {
	ID        uuid.UUID `json:"id" gorm:"primary_key;type:uuid;default:gen_random_uuid()"`
	Name      string    `json:"name" gorm:"size:255;not null;index"`
	Version   string    `json:"version" gorm:"size:128;not null;index"`
	Type      string    `json:"type" gorm:"size:50;not null;index"`
	Digest    string    `json:"digest" gorm:"size:255;not null;uniqueIndex"`
	MediaType string    `json:"media_type" gorm:"size:255;not null"`

	RepositoryID *uuid.UUID `json:"repository_id" gorm:"type:uuid;index"`
	Repository   Repository `json:"repository,omitempty" gorm:"foreignKey:RepositoryID"`

	Annotations map[string]string `json:"annotations,omitempty" gorm:"serializer:json;type:jsonb"`
	Size        int64             `json:"size"`

	Layers []Layer `json:"layers,omitempty" gorm:"many2many:artefact_layers;"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
