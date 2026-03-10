package repotype

import (
	"context"
	"io"
)

type TypeDescriptor struct {
	ArtifactType    string
	ConfigMediaType string
	LayerMediaType  string
}

// PushResult — returned by BeforePush. Nil fields keep original values.
type PushResult struct {
	Content          io.Reader
	ExtraAnnotations map[string]string
}

// PullResult — returned by AfterPull. Nil fields keep original values.
type PullResult struct {
	Content   io.ReadCloser
	MediaType string
}

// RepositoryTypePlugin is the contract for repository type plugins.
// "generic" is always built-in; others (docker, npm, helm...) are loaded from config.
type RepositoryTypePlugin interface {
	Name() string
	Description() string
	Descriptor() TypeDescriptor

	// BeforePush validates/transforms content before OCI push. Return nil for passthrough.
	BeforePush(ctx context.Context, name, version string, content io.Reader, annotations map[string]string) (*PushResult, error)

	// AfterPull transforms content after OCI pull. Return nil for passthrough.
	AfterPull(ctx context.Context, name, version string, content io.ReadCloser) (*PullResult, error)
}
