package models

import "github.com/google/uuid"

// Layer représente une couche (blob) OCI d'un artefact
type Layer struct {
	ID          uuid.UUID         `json:"id" gorm:"primary_key;type:uuid;default:gen_random_uuid()"`
	Digest      string            `json:"digest" gorm:"size:255;not null;index"`
	MediaType   string            `json:"media_type" gorm:"size:255;not null"`
	Size        int64             `json:"size"`
	Annotations map[string]string `json:"annotations,omitempty" gorm:"serializer:json;type:jsonb"`
}
