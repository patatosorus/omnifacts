package plugins

import (
	"context"
	"io"

	"omnifacts/internal/repotype"
)

type ociPlugin struct{}

func OCI() repotype.RepositoryTypePlugin {
	return &ociPlugin{}
}

func (p *ociPlugin) Name() string        { return "oci" }
func (p *ociPlugin) Description() string { return "Artefacts OCI natifs" }

func (p *ociPlugin) Descriptor() repotype.TypeDescriptor {
	return repotype.TypeDescriptor{
		ArtifactType:    "",
		ConfigMediaType: repotype.MediaTypeOCIConfig,
		LayerMediaType:  repotype.MediaTypeOCILayer,
	}
}

func (p *ociPlugin) BeforePush(_ context.Context, _, _ string, _ io.Reader, _ map[string]string) (*repotype.PushResult, error) {
	return nil, nil
}

func (p *ociPlugin) AfterPull(_ context.Context, _, _ string, _ io.ReadCloser) (*repotype.PullResult, error) {
	return nil, nil
}
