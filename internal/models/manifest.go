package models

import (
	"time"

	"github.com/google/uuid"
)

// OCIManifest représente un manifeste OCI stocké pour référence
type OCIManifest struct {
	ID           uuid.UUID `json:"id" gorm:"primary_key;type:uuid;default:gen_random_uuid()"`
	Digest       string    `json:"digest" gorm:"size:255;not null;uniqueIndex"`
	MediaType    string    `json:"media_type" gorm:"size:255;not null"`
	ArtifactType string    `json:"artifact_type" gorm:"size:255"`
	Content      []byte    `json:"-" gorm:"type:bytea"`
	Size         int64     `json:"size"`

	ArtefactID uuid.UUID `json:"artefact_id" gorm:"type:uuid;index"`

	CreatedAt time.Time `json:"created_at"`
}
