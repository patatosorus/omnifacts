package storage

import (
	"context"
	"io"
)

// StorageBackend définit l'interface de stockage et récupération du contenu des artefacts via OCI/ORAS
type StorageBackend interface {
	// PushArtefact pousse un artefact vers le registre OCI
	// repo: nom du dépôt OCI (ex: "omnifacts/mylib")
	// tag: référence/tag (ex: "1.0.0", "latest")
	// layerMediaType: type de média de la couche de contenu
	// artifactType: champ artifactType du manifeste OCI 1.1 (peut être vide)
	// content: contenu binaire de l'artefact
	// annotations: annotations OCI à attacher au manifeste
	PushArtefact(ctx context.Context, repo, tag, layerMediaType, artifactType string, content io.Reader, annotations map[string]string) (digest string, size int64, err error)

	// PullArtefact récupère le contenu d'un artefact depuis le registre OCI
	// repo: nom du dépôt OCI
	// reference: tag ou digest (ex: "1.0.0" ou "sha256:abc...")
	PullArtefact(ctx context.Context, repo, reference string) (io.ReadCloser, error)

	// DeleteArtefact supprime un artefact du registre OCI
	// repo: nom du dépôt OCI
	// reference: tag ou digest
	DeleteArtefact(ctx context.Context, repo, reference string) error

	// ResolveDigest résout une référence (tag) en digest
	ResolveDigest(ctx context.Context, repo, reference string) (string, error)
}
