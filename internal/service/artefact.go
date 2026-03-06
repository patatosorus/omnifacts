package service

import (
	"context"
	"fmt"
	"io"

	"omnifacts/internal/db"
	"omnifacts/internal/models"
	"omnifacts/internal/oci"
	"omnifacts/internal/storage"

	"github.com/google/uuid"
)

type ArtefactService interface {
	CreateArtefact(ctx context.Context, repoName, name, version string, content io.Reader, annotations map[string]string) (*models.Artefact, error)
	RetrieveArtefacts() ([]models.Artefact, error)
	RetrieveArtefactsByRepo(repoName string) ([]models.Artefact, error)
	RetrieveArtefactsByType(artefactType string) ([]models.Artefact, error)
	GetArtefactContent(ctx context.Context, id uuid.UUID) (io.ReadCloser, error)
	GetArtefactByNameAndVersion(name, version string) (*models.Artefact, error)
	UpdateArtefact(artefact *models.Artefact) error
	DeleteArtefact(ctx context.Context, id uuid.UUID) error
}

type artefactService struct {
	artefactDB db.ArtefactDB
	repoDB     db.RepositoryDB
	storage    storage.StorageBackend
	namespace  string
}

func NewArtefactService(artefactDB db.ArtefactDB, repoDB db.RepositoryDB, storage storage.StorageBackend, namespace string) ArtefactService {
	return &artefactService{
		artefactDB: artefactDB,
		repoDB:     repoDB,
		storage:    storage,
		namespace:  namespace,
	}
}

func (s *artefactService) CreateArtefact(ctx context.Context, repoName, name, version string, content io.Reader, annotations map[string]string) (*models.Artefact, error) {
	repo, err := s.repoDB.FindByName(repoName)
	if err != nil {
		return nil, fmt.Errorf("dépôt '%s' non trouvé : %w", repoName, err)
	}

	typeDesc, err := oci.GetTypeDescriptor(oci.ArtefactType(repo.ArtefactType))
	if err != nil {
		return nil, fmt.Errorf("type d'artefact invalide : %w", err)
	}

	ociRepoName := fmt.Sprintf("%s/%s/%s", s.namespace, repoName, name)

	digest, size, err := s.storage.PushArtefact(
		ctx,
		ociRepoName,
		version,
		typeDesc.LayerMediaType,
		typeDesc.ArtifactType,
		content,
		annotations,
	)
	if err != nil {
		return nil, fmt.Errorf("impossible de pousser l'artefact : %w", err)
	}

	artefact := &models.Artefact{
		Name:         name,
		Version:      version,
		Type:         repo.ArtefactType,
		Digest:       digest,
		MediaType:    typeDesc.LayerMediaType,
		Annotations:  annotations,
		Size:         size,
		RepositoryID: &repo.ID,
	}

	if err := s.artefactDB.Create(artefact); err != nil {
		_ = s.storage.DeleteArtefact(ctx, ociRepoName, digest)
		return nil, fmt.Errorf("impossible de sauvegarder les métadonnées : %w", err)
	}

	return artefact, nil
}

func (s *artefactService) RetrieveArtefacts() ([]models.Artefact, error) {
	return s.artefactDB.Retrieve()
}

func (s *artefactService) RetrieveArtefactsByRepo(repoName string) ([]models.Artefact, error) {
	repo, err := s.repoDB.FindByName(repoName)
	if err != nil {
		return nil, fmt.Errorf("dépôt '%s' non trouvé : %w", repoName, err)
	}
	return s.artefactDB.ListByRepo(repo.ID)
}

func (s *artefactService) RetrieveArtefactsByType(artefactType string) ([]models.Artefact, error) {
	return s.artefactDB.ListByType(artefactType)
}

func (s *artefactService) GetArtefactContent(ctx context.Context, id uuid.UUID) (io.ReadCloser, error) {
	artefact, err := s.artefactDB.Find(id)
	if err != nil {
		return nil, fmt.Errorf("artefact non trouvé : %w", err)
	}

	var repoPrefix string
	if artefact.RepositoryID != nil {
		repo, err := s.repoDB.FindByID(*artefact.RepositoryID)
		if err == nil {
			repoPrefix = repo.Name + "/"
		}
	}

	ociRepoName := fmt.Sprintf("%s/%s%s", s.namespace, repoPrefix, artefact.Name)
	return s.storage.PullArtefact(ctx, ociRepoName, artefact.Digest)
}

func (s *artefactService) GetArtefactByNameAndVersion(name, version string) (*models.Artefact, error) {
	return s.artefactDB.FindByNameAndVersion(name, version)
}

func (s *artefactService) UpdateArtefact(artefact *models.Artefact) error {
	return s.artefactDB.Update(artefact)
}

func (s *artefactService) DeleteArtefact(ctx context.Context, id uuid.UUID) error {
	artefact, err := s.artefactDB.Find(id)
	if err != nil {
		return fmt.Errorf("artefact non trouvé : %w", err)
	}

	var repoPrefix string
	if artefact.RepositoryID != nil {
		repo, err := s.repoDB.FindByID(*artefact.RepositoryID)
		if err == nil {
			repoPrefix = repo.Name + "/"
		}
	}

	ociRepoName := fmt.Sprintf("%s/%s%s", s.namespace, repoPrefix, artefact.Name)
	if err := s.storage.DeleteArtefact(ctx, ociRepoName, artefact.Digest); err != nil {
		return fmt.Errorf("impossible de supprimer du registre : %w", err)
	}

	return s.artefactDB.Delete(id)
}
