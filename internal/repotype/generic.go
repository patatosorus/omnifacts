package repotype

import (
	"context"
	"io"
)

const (
	MediaTypeOCIEmptyJSON   = "application/vnd.oci.empty.v1+json"
	MediaTypeGenericContent = "application/octet-stream"
)

type genericPlugin struct{}

func NewGenericPlugin() RepositoryTypePlugin {
	return &genericPlugin{}
}

func (p *genericPlugin) Name() string {
	return GenericTypeName
}

func (p *genericPlugin) Description() string {
	return "Dépôt générique pour artefacts binaires quelconques"
}

func (p *genericPlugin) Descriptor() TypeDescriptor {
	return TypeDescriptor{
		ArtifactType:    "",
		ConfigMediaType: MediaTypeOCIEmptyJSON,
		LayerMediaType:  MediaTypeGenericContent,
	}
}

func (p *genericPlugin) BeforePush(_ context.Context, _, _ string, _ io.Reader, _ map[string]string) (*PushResult, error) {
	return nil, nil
}

func (p *genericPlugin) AfterPull(_ context.Context, _, _ string, _ io.ReadCloser) (*PullResult, error) {
	return nil, nil
}
