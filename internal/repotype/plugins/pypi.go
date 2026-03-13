package plugins

import (
	"context"
	"io"

	"omnifacts/internal/repotype"
)

type pypiPlugin struct{}

func PyPI() repotype.RepositoryTypePlugin {
	return &pypiPlugin{}
}

func (p *pypiPlugin) Name() string        { return "pypi" }
func (p *pypiPlugin) Description() string { return "Paquets Python (PyPI)" }

func (p *pypiPlugin) Descriptor() repotype.TypeDescriptor {
	return repotype.TypeDescriptor{
		ArtifactType:    repotype.MediaTypePyPIPackage,
		ConfigMediaType: repotype.MediaTypeOCIEmptyJSON,
		LayerMediaType:  repotype.MediaTypePyPIWheel,
	}
}

func (p *pypiPlugin) BeforePush(_ context.Context, _, _ string, _ io.Reader, _ map[string]string) (*repotype.PushResult, error) {
	return nil, nil
}

func (p *pypiPlugin) AfterPull(_ context.Context, _, _ string, _ io.ReadCloser) (*repotype.PullResult, error) {
	return nil, nil
}
