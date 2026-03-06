package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/memory"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
)

// ORASStorage implémente StorageBackend en utilisant ORAS pour pousser/tirer vers un registre OCI
type ORASStorage struct {
	registryURL string
	plainHTTP   bool
}

// NewORASStorage crée un nouveau backend de stockage ORAS
func NewORASStorage(registryURL string, plainHTTP bool) *ORASStorage {
	return &ORASStorage{
		registryURL: registryURL,
		plainHTTP:   plainHTTP,
	}
}

func (s *ORASStorage) newRepository(repoName string) (*remote.Repository, error) {
	ref := fmt.Sprintf("%s/%s", s.registryURL, repoName)
	repo, err := remote.NewRepository(ref)
	if err != nil {
		return nil, fmt.Errorf("impossible de créer la référence du dépôt : %w", err)
	}
	repo.PlainHTTP = s.plainHTTP
	repo.Client = &auth.Client{
		Client: retry.DefaultClient,
		Cache:  auth.NewCache(),
	}
	return repo, nil
}

func (s *ORASStorage) PushArtefact(ctx context.Context, repoName, tag, layerMediaType, artifactType string, content io.Reader, annotations map[string]string) (string, int64, error) {
	repo, err := s.newRepository(repoName)
	if err != nil {
		return "", 0, err
	}

	contentBytes, err := io.ReadAll(content)
	if err != nil {
		return "", 0, fmt.Errorf("impossible de lire le contenu : %w", err)
	}

	memStore := memory.New()

	layerDesc, err := oras.PushBytes(ctx, memStore, layerMediaType, contentBytes)
	if err != nil {
		return "", 0, fmt.Errorf("impossible de pousser le blob : %w", err)
	}

	packOpts := oras.PackManifestOptions{
		Layers:              []ocispec.Descriptor{layerDesc},
		ManifestAnnotations: annotations,
	}

	manifestDesc, err := oras.PackManifest(ctx, memStore, oras.PackManifestVersion1_1, artifactType, packOpts)
	if err != nil {
		return "", 0, fmt.Errorf("impossible de créer le manifeste : %w", err)
	}

	if err := memStore.Tag(ctx, manifestDesc, tag); err != nil {
		return "", 0, fmt.Errorf("impossible de taguer le manifeste : %w", err)
	}

	manifestDesc, err = oras.Copy(ctx, memStore, tag, repo, tag, oras.DefaultCopyOptions)
	if err != nil {
		return "", 0, fmt.Errorf("impossible de pousser vers le registre : %w", err)
	}

	slog.Info("Artefact poussé vers le registre",
		"repo", repoName,
		"tag", tag,
		"digest", manifestDesc.Digest.String(),
		"size", manifestDesc.Size,
	)

	return manifestDesc.Digest.String(), manifestDesc.Size, nil
}

func (s *ORASStorage) PullArtefact(ctx context.Context, repoName, reference string) (io.ReadCloser, error) {
	repo, err := s.newRepository(repoName)
	if err != nil {
		return nil, err
	}

	memStore := memory.New()

	desc, err := oras.Copy(ctx, repo, reference, memStore, reference, oras.DefaultCopyOptions)
	if err != nil {
		return nil, fmt.Errorf("impossible de tirer l'artefact : %w", err)
	}

	manifestContent, err := memStore.Fetch(ctx, desc)
	if err != nil {
		return nil, fmt.Errorf("impossible de lire le manifeste : %w", err)
	}
	defer manifestContent.Close()

	manifestBytes, err := io.ReadAll(manifestContent)
	if err != nil {
		return nil, fmt.Errorf("impossible de lire le contenu du manifeste : %w", err)
	}

	var manifest ocispec.Manifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return nil, fmt.Errorf("impossible de parser le manifeste : %w", err)
	}

	if len(manifest.Layers) == 0 {
		return nil, fmt.Errorf("aucune couche trouvée dans le manifeste")
	}

	layerContent, err := memStore.Fetch(ctx, manifest.Layers[0])
	if err != nil {
		return nil, fmt.Errorf("impossible de lire la couche : %w", err)
	}

	return layerContent, nil
}

func (s *ORASStorage) DeleteArtefact(ctx context.Context, repoName, reference string) error {
	repo, err := s.newRepository(repoName)
	if err != nil {
		return err
	}

	desc, err := repo.Resolve(ctx, reference)
	if err != nil {
		return fmt.Errorf("impossible de résoudre la référence : %w", err)
	}

	if err := repo.Delete(ctx, desc); err != nil {
		return fmt.Errorf("impossible de supprimer l'artefact : %w", err)
	}

	slog.Info("Artefact supprimé du registre", "repo", repoName, "reference", reference)
	return nil
}

func (s *ORASStorage) ResolveDigest(ctx context.Context, repoName, reference string) (string, error) {
	repo, err := s.newRepository(repoName)
	if err != nil {
		return "", err
	}

	desc, err := repo.Resolve(ctx, reference)
	if err != nil {
		return "", fmt.Errorf("impossible de résoudre la référence : %w", err)
	}

	return desc.Digest.String(), nil
}


