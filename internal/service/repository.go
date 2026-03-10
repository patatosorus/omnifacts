package service

import (
	"fmt"

	"omnifacts/internal/db"
	"omnifacts/internal/models"
	"omnifacts/internal/repotype"

	"github.com/google/uuid"
)

type RepositoryService interface {
	Create(name, description, artefactType string, storageMode models.StorageMode, upstreamURL string) (*models.Repository, error)
	FindByID(id uuid.UUID) (*models.Repository, error)
	FindByName(name string) (*models.Repository, error)
	List() ([]models.Repository, error)
	Update(repo *models.Repository) error
	Delete(id uuid.UUID) error

	// Gestion des permissions
	SetPermission(userID, repoID uuid.UUID, level models.PermissionLevel) error
	GetPermission(userID, repoID uuid.UUID) (*models.Permission, error)
	ListPermissions(repoID uuid.UUID) ([]models.Permission, error)
	RemovePermission(userID, repoID uuid.UUID) error
}

type repositoryService struct {
	repoDB       db.RepositoryDB
	permissionDB db.PermissionDB
	registry     *repotype.Registry
}

func NewRepositoryService(repoDB db.RepositoryDB, permissionDB db.PermissionDB, registry *repotype.Registry) RepositoryService {
	return &repositoryService{repoDB: repoDB, permissionDB: permissionDB, registry: registry}
}

func (s *repositoryService) Create(name, description, artefactType string, storageMode models.StorageMode, upstreamURL string) (*models.Repository, error) {
	if !s.registry.IsRegistered(artefactType) {
		return nil, fmt.Errorf("type de dépôt '%s' non disponible (plugin non chargé)", artefactType)
	}

	if storageMode == models.StorageMirror || storageMode == models.StorageRemote {
		if upstreamURL == "" {
			return nil, fmt.Errorf("l'URL upstream est requise pour le mode %s", storageMode)
		}
	}

	repo := &models.Repository{
		Name:         name,
		Description:  description,
		ArtefactType: artefactType,
		StorageMode:  storageMode,
		UpstreamURL:  upstreamURL,
	}

	if err := s.repoDB.Create(repo); err != nil {
		return nil, fmt.Errorf("impossible de créer le dépôt : %w", err)
	}

	return repo, nil
}

func (s *repositoryService) FindByID(id uuid.UUID) (*models.Repository, error) {
	return s.repoDB.FindByID(id)
}

func (s *repositoryService) FindByName(name string) (*models.Repository, error) {
	return s.repoDB.FindByName(name)
}

func (s *repositoryService) List() ([]models.Repository, error) {
	return s.repoDB.List()
}

func (s *repositoryService) Update(repo *models.Repository) error {
	return s.repoDB.Update(repo)
}

func (s *repositoryService) Delete(id uuid.UUID) error {
	perms, err := s.permissionDB.ListByRepo(id)
	if err == nil {
		for _, p := range perms {
			_ = s.permissionDB.Delete(p.ID)
		}
	}
	return s.repoDB.Delete(id)
}

func (s *repositoryService) SetPermission(userID, repoID uuid.UUID, level models.PermissionLevel) error {
	existing, err := s.permissionDB.FindByUserAndRepo(userID, repoID)
	if err == nil {
		existing.Level = level
		return s.permissionDB.Update(existing)
	}

	perm := &models.Permission{
		UserID:       userID,
		RepositoryID: repoID,
		Level:        level,
	}
	return s.permissionDB.Create(perm)
}

func (s *repositoryService) GetPermission(userID, repoID uuid.UUID) (*models.Permission, error) {
	return s.permissionDB.FindByUserAndRepo(userID, repoID)
}

func (s *repositoryService) ListPermissions(repoID uuid.UUID) ([]models.Permission, error) {
	return s.permissionDB.ListByRepo(repoID)
}

func (s *repositoryService) RemovePermission(userID, repoID uuid.UUID) error {
	return s.permissionDB.DeleteByUserAndRepo(userID, repoID)
}
