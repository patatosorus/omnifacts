package plugins

import (
	"context"
	"io"

	"omnifacts/internal/repotype"
)

type terraformPlugin struct{}

func Terraform() repotype.RepositoryTypePlugin {
	return &terraformPlugin{}
}

func (p *terraformPlugin) Name() string        { return "terraform" }
func (p *terraformPlugin) Description() string { return "Modules Terraform / OpenTofu" }

func (p *terraformPlugin) Descriptor() repotype.TypeDescriptor {
	return repotype.TypeDescriptor{
		ArtifactType:    repotype.MediaTypeTerraformModule,
		ConfigMediaType: repotype.MediaTypeOCIEmptyJSON,
		LayerMediaType:  repotype.MediaTypeTerraformModuleLayer,
	}
}

func (p *terraformPlugin) BeforePush(_ context.Context, _, _ string, _ io.Reader, _ map[string]string) (*repotype.PushResult, error) {
	return nil, nil
}

func (p *terraformPlugin) AfterPull(_ context.Context, _, _ string, _ io.ReadCloser) (*repotype.PullResult, error) {
	return nil, nil
}
