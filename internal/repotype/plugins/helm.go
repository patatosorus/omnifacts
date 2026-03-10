package plugins

import (
	"context"
	"io"

	"omnifacts/internal/repotype"
)

type helmPlugin struct{}

func Helm() repotype.RepositoryTypePlugin {
	return &helmPlugin{}
}

func (p *helmPlugin) Name() string        { return "helm" }
func (p *helmPlugin) Description() string { return "Charts Helm pour Kubernetes" }

func (p *helmPlugin) Descriptor() repotype.TypeDescriptor {
	return repotype.TypeDescriptor{
		ArtifactType:    "",
		ConfigMediaType: repotype.MediaTypeHelmConfig,
		LayerMediaType:  repotype.MediaTypeHelmChart,
	}
}

func (p *helmPlugin) BeforePush(_ context.Context, _, _ string, _ io.Reader, _ map[string]string) (*repotype.PushResult, error) {
	return nil, nil
}

func (p *helmPlugin) AfterPull(_ context.Context, _, _ string, _ io.ReadCloser) (*repotype.PullResult, error) {
	return nil, nil
}
