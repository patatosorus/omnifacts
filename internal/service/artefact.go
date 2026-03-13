package service

import (
	"context"
	"fmt"
	"io"

	"omnifacts/internal/db"
	"omnifacts/internal/models"
	"omnifacts/internal/repotype"
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
	registry   *repotype.Registry
}

func NewArtefactService(artefactDB db.ArtefactDB, repoDB db.RepositoryDB, storage storage.StorageBackend, namespace string, registry *repotype.Registry) ArtefactService {
	return &artefactService{
		artefactDB: artefactDB,
		repoDB:     repoDB,
		storage:    storage,
		namespace:  namespace,
		registry:   registry,
	}
}

func (s *artefactService) CreateArtefact(ctx context.Context, repoName, name, version string, content io.Reader, annotations map[string]string) (*models.Artefact, error) {
	repo, err := s.repoDB.FindByName(repoName)
	if err != nil {
		return nil, fmt.Errorf("dépôt '%s' non trouvé : %w", repoName, err)
	}

	plugin, err := s.resolvePlugin(repo.ArtefactType)
	if err != nil {
		return nil, err
	}

	if !s.registry.IsRegistered(repo.ArtefactType) {
		return nil, fmt.Errorf("le dépôt '%s' est en lecture seule (plugin '%s' non chargé)", repoName, repo.ArtefactType)
	}

	typeDesc := plugin.Descriptor()

	pushContent := content
	if result, err := plugin.BeforePush(ctx, name, version, content, annotations); err != nil {
		return nil, fmt.Errorf("erreur plugin avant push : %w", err)
	} else if result != nil {
		if result.Content != nil {
			pushContent = result.Content
		}
		if result.ExtraAnnotations != nil {
			if annotations == nil {
				annotations = make(map[string]string)
			}
			for k, v := range result.ExtraAnnotations {
				annotations[k] = v
			}
		}
	}

	ociRepoName := fmt.Sprintf("%s/%s/%s", s.namespace, repoName, name)

	digest, size, err := s.storage.PushArtefact(
		ctx,
		ociRepoName,
		version,
		typeDesc.LayerMediaType,
		typeDesc.ArtifactType,
		pushContent,
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

	var repoName string
	var repoType string
	if artefact.RepositoryID != nil {
		repo, err := s.repoDB.FindByID(*artefact.RepositoryID)
		if err == nil {
			repoName = repo.Name
			repoType = repo.ArtefactType
		}
	}

	var repoPrefix string
	if repoName != "" {
		repoPrefix = repoName + "/"
	}

	ociRepoName := fmt.Sprintf("%s/%s%s", s.namespace, repoPrefix, artefact.Name)
	content, err := s.storage.PullArtefact(ctx, ociRepoName, artefact.Digest)
	if err != nil {
		return nil, err
	}

	plugin, resolveErr := s.resolvePlugin(repoType)
	if resolveErr == nil {
		if result, err := plugin.AfterPull(ctx, artefact.Name, artefact.Version, content); err != nil {
			content.Close()
			return nil, fmt.Errorf("erreur plugin après pull : %w", err)
		} else if result != nil && result.Content != nil {
			content.Close()
			return result.Content, nil
		}
	}

	return content, nil
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

	if artefact.RepositoryID != nil {
		repo, err := s.repoDB.FindByID(*artefact.RepositoryID)
		if err == nil && !s.registry.IsRegistered(repo.ArtefactType) {
			return fmt.Errorf("le dépôt '%s' est en lecture seule (plugin '%s' non chargé)", repo.Name, repo.ArtefactType)
		}
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

func (s *artefactService) resolvePlugin(artefactType string) (repotype.RepositoryTypePlugin, error) {
	plugin, err := s.registry.Get(artefactType)
	if err != nil {
		return s.registry.GetGeneric(), nil
	}
	return plugin, nil
}
