package oci

import "fmt"

// ArtefactType représente le type d'un artefact géré par le registre
type ArtefactType string

const (
	TypeDocker    ArtefactType = "docker"
	TypeOCI       ArtefactType = "oci"
	TypeHelm      ArtefactType = "helm"
	TypePyPI      ArtefactType = "pypi"
	TypeNpm       ArtefactType = "npm"
	TypeTerraform ArtefactType = "terraform"
	TypeMaven     ArtefactType = "maven"
	TypeGeneric   ArtefactType = "generic"
)

// TypeDescriptor décrit le mapping OCI pour un type d'artefact
type TypeDescriptor struct {
	ArtifactType    string // Champ artifactType du manifeste OCI 1.1
	ConfigMediaType string // Type de média pour le blob de configuration
	LayerMediaType  string // Type de média par défaut pour les couches de contenu
}

// TypeRegistry associe chaque ArtefactType à son TypeDescriptor OCI
var TypeRegistry = map[ArtefactType]TypeDescriptor{
	TypeDocker: {
		ArtifactType:    "",
		ConfigMediaType: MediaTypeDockerConfig,
		LayerMediaType:  MediaTypeDockerLayer,
	},
	TypeOCI: {
		ArtifactType:    "",
		ConfigMediaType: MediaTypeOCIConfig,
		LayerMediaType:  MediaTypeOCILayer,
	},
	TypeHelm: {
		ArtifactType:    "",
		ConfigMediaType: MediaTypeHelmConfig,
		LayerMediaType:  MediaTypeHelmChart,
	},
	TypeTerraform: {
		ArtifactType:    MediaTypeTerraformModule,
		ConfigMediaType: MediaTypeOCIEmptyJSON,
		LayerMediaType:  MediaTypeTerraformModuleLayer,
	},
	TypePyPI: {
		ArtifactType:    MediaTypePyPIPackage,
		ConfigMediaType: MediaTypeOCIEmptyJSON,
		LayerMediaType:  MediaTypePyPIWheel,
	},
	TypeNpm: {
		ArtifactType:    MediaTypeNpmPackage,
		ConfigMediaType: MediaTypeOCIEmptyJSON,
		LayerMediaType:  MediaTypeNpmTarball,
	},
	TypeMaven: {
		ArtifactType:    MediaTypeMavenArtifact,
		ConfigMediaType: MediaTypeOCIEmptyJSON,
		LayerMediaType:  MediaTypeMavenArtifact,
	},
	TypeGeneric: {
		ArtifactType:    "",
		ConfigMediaType: MediaTypeOCIEmptyJSON,
		LayerMediaType:  MediaTypeGenericContent,
	},
}

// GetTypeDescriptor retourne le descripteur pour un type d'artefact donné
func GetTypeDescriptor(t ArtefactType) (TypeDescriptor, error) {
	desc, ok := TypeRegistry[t]
	if !ok {
		return TypeDescriptor{}, fmt.Errorf("type d'artefact inconnu : %s", t)
	}
	return desc, nil
}

// ParseArtefactType valide et convertit une chaîne en ArtefactType
func ParseArtefactType(s string) (ArtefactType, error) {
	t := ArtefactType(s)
	if _, ok := TypeRegistry[t]; !ok {
		return "", fmt.Errorf("type d'artefact invalide : %s", s)
	}
	return t, nil
}

// ValidArtefactTypes retourne la liste de tous les types d'artefacts supportés
func ValidArtefactTypes() []ArtefactType {
	types := make([]ArtefactType, 0, len(TypeRegistry))
	for t := range TypeRegistry {
		types = append(types, t)
	}
	return types
}
