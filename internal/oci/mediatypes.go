package oci

// Types de médias OCI standard
const (
	MediaTypeOCIManifest  = "application/vnd.oci.image.manifest.v1+json"
	MediaTypeOCIIndex     = "application/vnd.oci.image.index.v1+json"
	MediaTypeOCIConfig    = "application/vnd.oci.image.config.v1+json"
	MediaTypeOCILayer     = "application/vnd.oci.image.layer.v1.tar+gzip"
	MediaTypeOCIEmptyJSON = "application/vnd.oci.empty.v1+json"
)

// Types de médias Docker
const (
	MediaTypeDockerManifest = "application/vnd.docker.distribution.manifest.v2+json"
	MediaTypeDockerConfig   = "application/vnd.docker.container.image.v1+json"
	MediaTypeDockerLayer    = "application/vnd.docker.image.rootfs.diff.tar.gzip"
)

// Types de médias Helm
const (
	MediaTypeHelmConfig     = "application/vnd.cncf.helm.config.v1+json"
	MediaTypeHelmChart      = "application/vnd.cncf.helm.chart.content.v1.tar+gzip"
	MediaTypeHelmProvenance = "application/vnd.cncf.helm.chart.provenance.v1.prov"
)

// Types de médias Terraform / OpenTofu
const (
	MediaTypeTerraformModule      = "application/vnd.opentofu.modulepkg"
	MediaTypeTerraformModuleLayer = "archive/zip"
)

// Types de médias PyPI
const (
	MediaTypePyPIPackage = "application/vnd.pypi.package.v1"
	MediaTypePyPIWheel   = "application/vnd.pypi.package.v1+gzip"
)

// Types de médias npm
const (
	MediaTypeNpmPackage = "application/vnd.npm.package.v1"
	MediaTypeNpmTarball = "application/vnd.npm.package.v1+gzip"
)

// Types de médias Maven
const (
	MediaTypeMavenArtifact = "application/vnd.maven.artifact.v1"
)

// Type de média générique
const (
	MediaTypeGenericContent = "application/octet-stream"
)

// Configuration vide OCI (utilisée pour les artefacts non-container)
const (
	EmptyConfigDigest  = "sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a"
	EmptyConfigContent = "{}"
)

const EmptyConfigSize int64 = 2
