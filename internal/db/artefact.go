package db

import (
	"omnifacts/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ArtefactDB définit l'interface de gestion des artefacts en base de données
type ArtefactDB interface {
	Create(artefact *models.Artefact) error
	Retrieve() ([]models.Artefact, error)
	Find(id uuid.UUID) (*models.Artefact, error)
	FindByDigest(digest string) (*models.Artefact, error)
	FindByNameAndVersion(name, version string) (*models.Artefact, error)
	ListByType(artefactType string) ([]models.Artefact, error)
	ListByRepo(repoID uuid.UUID) ([]models.Artefact, error)
	Update(artefact *models.Artefact) error
	Delete(id uuid.UUID) error
}

type artefactDB struct {
	db *gorm.DB
}

func NewArtefactDB(db *gorm.DB) ArtefactDB {
	return &artefactDB{db: db}
}

func (r *artefactDB) Create(artefact *models.Artefact) error {
	return r.db.Create(artefact).Error
}

func (r *artefactDB) Retrieve() ([]models.Artefact, error) {
	var artefacts []models.Artefact
	err := r.db.Preload("Layers").Find(&artefacts).Error
	return artefacts, err
}

func (r *artefactDB) Find(id uuid.UUID) (*models.Artefact, error) {
	var artefact models.Artefact
	err := r.db.Preload("Layers").First(&artefact, "id = ?", id).Error
	return &artefact, err
}

func (r *artefactDB) FindByDigest(digest string) (*models.Artefact, error) {
	var artefact models.Artefact
	err := r.db.Preload("Layers").First(&artefact, "digest = ?", digest).Error
	return &artefact, err
}

func (r *artefactDB) FindByNameAndVersion(name, version string) (*models.Artefact, error) {
	var artefact models.Artefact
	err := r.db.Preload("Layers").First(&artefact, "name = ? AND version = ?", name, version).Error
	return &artefact, err
}

func (r *artefactDB) ListByType(artefactType string) ([]models.Artefact, error) {
	var artefacts []models.Artefact
	err := r.db.Preload("Layers").Where("type = ?", artefactType).Find(&artefacts).Error
	return artefacts, err
}

func (r *artefactDB) ListByRepo(repoID uuid.UUID) ([]models.Artefact, error) {
	var artefacts []models.Artefact
	err := r.db.Preload("Layers").Where("repository_id = ?", repoID).Find(&artefacts).Error
	return artefacts, err
}

func (r *artefactDB) Update(artefact *models.Artefact) error {
	return r.db.Save(artefact).Error
}

func (r *artefactDB) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Artefact{}, id).Error
}
