package repotype

const (
	MediaTypeOCIManifest = "application/vnd.oci.image.manifest.v1+json"
	MediaTypeOCIIndex    = "application/vnd.oci.image.index.v1+json"
	MediaTypeOCIConfig   = "application/vnd.oci.image.config.v1+json"
	MediaTypeOCILayer    = "application/vnd.oci.image.layer.v1.tar+gzip"

	MediaTypeDockerManifest = "application/vnd.docker.distribution.manifest.v2+json"
	MediaTypeDockerConfig   = "application/vnd.docker.container.image.v1+json"
	MediaTypeDockerLayer    = "application/vnd.docker.image.rootfs.diff.tar.gzip"

	MediaTypeHelmConfig     = "application/vnd.cncf.helm.config.v1+json"
	MediaTypeHelmChart      = "application/vnd.cncf.helm.chart.content.v1.tar+gzip"
	MediaTypeHelmProvenance = "application/vnd.cncf.helm.chart.provenance.v1.prov"

	MediaTypeTerraformModule      = "application/vnd.opentofu.modulepkg"
	MediaTypeTerraformModuleLayer = "archive/zip"

	MediaTypePyPIPackage = "application/vnd.pypi.package.v1"
	MediaTypePyPIWheel   = "application/vnd.pypi.package.v1+gzip"

	MediaTypeNpmPackage = "application/vnd.npm.package.v1"
	MediaTypeNpmTarball = "application/vnd.npm.package.v1+gzip"

	MediaTypeMavenArtifact = "application/vnd.maven.artifact.v1"

	EmptyConfigDigest        = "sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a"
	EmptyConfigContent       = "{}"
	EmptyConfigSize    int64 = 2
)
