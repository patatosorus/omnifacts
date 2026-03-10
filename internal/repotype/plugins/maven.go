package plugins

import (
	"context"
	"io"

	"omnifacts/internal/repotype"
)

type mavenPlugin struct{}

func Maven() repotype.RepositoryTypePlugin {
	return &mavenPlugin{}
}

func (p *mavenPlugin) Name() string        { return "maven" }
func (p *mavenPlugin) Description() string { return "Artefacts Maven (JAR, WAR, POM)" }

func (p *mavenPlugin) Descriptor() repotype.TypeDescriptor {
	return repotype.TypeDescriptor{
		ArtifactType:    repotype.MediaTypeMavenArtifact,
		ConfigMediaType: repotype.MediaTypeOCIEmptyJSON,
		LayerMediaType:  repotype.MediaTypeMavenArtifact,
	}
}

func (p *mavenPlugin) BeforePush(_ context.Context, _, _ string, _ io.Reader, _ map[string]string) (*repotype.PushResult, error) {
	return nil, nil
}

func (p *mavenPlugin) AfterPull(_ context.Context, _, _ string, _ io.ReadCloser) (*repotype.PullResult, error) {
	return nil, nil
}
