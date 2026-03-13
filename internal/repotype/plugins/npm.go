package plugins

import (
	"context"
	"io"

	"omnifacts/internal/repotype"
)

type npmPlugin struct{}

func Npm() repotype.RepositoryTypePlugin {
	return &npmPlugin{}
}

func (p *npmPlugin) Name() string        { return "npm" }
func (p *npmPlugin) Description() string { return "Paquets npm pour Node.js" }

func (p *npmPlugin) Descriptor() repotype.TypeDescriptor {
	return repotype.TypeDescriptor{
		ArtifactType:    repotype.MediaTypeNpmPackage,
		ConfigMediaType: repotype.MediaTypeOCIEmptyJSON,
		LayerMediaType:  repotype.MediaTypeNpmTarball,
	}
}

func (p *npmPlugin) BeforePush(_ context.Context, _, _ string, _ io.Reader, _ map[string]string) (*repotype.PushResult, error) {
	return nil, nil
}

func (p *npmPlugin) AfterPull(_ context.Context, _, _ string, _ io.ReadCloser) (*repotype.PullResult, error) {
	return nil, nil
}
