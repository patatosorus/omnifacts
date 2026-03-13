package db

import (
	"omnifacts/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RepositoryDB interface {
	Create(repo *models.Repository) error
	FindByID(id uuid.UUID) (*models.Repository, error)
	FindByName(name string) (*models.Repository, error)
	Update(repo *models.Repository) error
	Delete(id uuid.UUID) error
	List() ([]models.Repository, error)
	ListByType(artefactType string) ([]models.Repository, error)
}

type repositoryDB struct {
	db *gorm.DB
}

func NewRepositoryDB(db *gorm.DB) RepositoryDB {
	return &repositoryDB{db: db}
}

func (r *repositoryDB) Create(repo *models.Repository) error {
	return r.db.Create(repo).Error
}

func (r *repositoryDB) FindByID(id uuid.UUID) (*models.Repository, error) {
	var repo models.Repository
	err := r.db.First(&repo, "id = ?", id).Error
	return &repo, err
}

func (r *repositoryDB) FindByName(name string) (*models.Repository, error) {
	var repo models.Repository
	err := r.db.First(&repo, "name = ?", name).Error
	return &repo, err
}

func (r *repositoryDB) Update(repo *models.Repository) error {
	return r.db.Save(repo).Error
}

func (r *repositoryDB) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Repository{}, id).Error
}

func (r *repositoryDB) List() ([]models.Repository, error) {
	var repos []models.Repository
	err := r.db.Find(&repos).Error
	return repos, err
}

func (r *repositoryDB) ListByType(artefactType string) ([]models.Repository, error) {
	var repos []models.Repository
	err := r.db.Where("artefact_type = ?", artefactType).Find(&repos).Error
	return repos, err
}
