package repotype

import (
	"fmt"
	"log/slog"
	"sync"
)

const GenericTypeName = "generic"

type Registry struct {
	mu      sync.RWMutex
	plugins map[string]RepositoryTypePlugin
}

func NewRegistry() *Registry {
	r := &Registry{
		plugins: make(map[string]RepositoryTypePlugin),
	}
	r.Register(NewGenericPlugin())
	return r
}

func (r *Registry) Register(plugin RepositoryTypePlugin) {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := plugin.Name()
	if _, exists := r.plugins[name]; exists {
		slog.Warn("Plugin de type de dépôt déjà enregistré, remplacement", "type", name)
	}
	r.plugins[name] = plugin
	slog.Info("Plugin de type de dépôt enregistré", "type", name, "description", plugin.Description())
}

func (r *Registry) Get(name string) (RepositoryTypePlugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugin, ok := r.plugins[name]
	if !ok {
		return nil, fmt.Errorf("type de dépôt non enregistré : %s", name)
	}
	return plugin, nil
}

func (r *Registry) IsRegistered(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.plugins[name]
	return ok
}

func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]string, 0, len(r.plugins))
	for name := range r.plugins {
		types = append(types, name)
	}
	return types
}

func (r *Registry) GetGeneric() RepositoryTypePlugin {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.plugins[GenericTypeName]
}
