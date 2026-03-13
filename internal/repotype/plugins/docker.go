package plugins

import (
	"context"
	"io"

	"omnifacts/internal/repotype"
)

type dockerPlugin struct{}

func Docker() repotype.RepositoryTypePlugin {
	return &dockerPlugin{}
}

func (p *dockerPlugin) Name() string        { return "docker" }
func (p *dockerPlugin) Description() string { return "Images Docker / OCI containers" }

func (p *dockerPlugin) Descriptor() repotype.TypeDescriptor {
	return repotype.TypeDescriptor{
		ArtifactType:    "",
		ConfigMediaType: repotype.MediaTypeDockerConfig,
		LayerMediaType:  repotype.MediaTypeDockerLayer,
	}
}

func (p *dockerPlugin) BeforePush(_ context.Context, _, _ string, _ io.Reader, _ map[string]string) (*repotype.PushResult, error) {
	return nil, nil
}

func (p *dockerPlugin) AfterPull(_ context.Context, _, _ string, _ io.ReadCloser) (*repotype.PullResult, error) {
	return nil, nil
}
