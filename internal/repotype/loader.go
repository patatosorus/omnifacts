package repotype

import (
	"fmt"
	"log/slog"
)

type PluginFactory func() RepositoryTypePlugin

var builtinFactories = map[string]PluginFactory{}

func RegisterBuiltinFactory(name string, factory PluginFactory) {
	builtinFactories[name] = factory
}

func LoadPlugins(registry *Registry, enabledNames []string) error {
	for _, name := range enabledNames {
		if name == GenericTypeName {
			continue
		}

		factory, ok := builtinFactories[name]
		if !ok {
			slog.Warn("Plugin non trouvé, ignoré", "type", name)
			return fmt.Errorf("plugin de type '%s' non trouvé dans les plugins intégrés", name)
		}

		registry.Register(factory())
	}
	return nil
}
