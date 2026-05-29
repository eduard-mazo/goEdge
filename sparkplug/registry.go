package sparkplug

import (
	"sync"
)

// Registry manages metric name-to-alias mappings for a Node or Device.
type Registry struct {
	mu      sync.RWMutex
	aliases map[string]uint64
	names   map[uint64]string
}

// NewRegistry creates a new empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		aliases: make(map[string]uint64),
		names:   make(map[uint64]string),
	}
}

// Register assigns or retrieves an alias for a metric name.
func (r *Registry) Register(name string, alias uint64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.aliases[name] = alias
	r.names[alias] = name
}

// GetAlias returns the alias for a given name.
func (r *Registry) GetAlias(name string) (uint64, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.aliases[name]
	return a, ok
}

// GetName returns the name for a given alias.
func (r *Registry) GetName(alias uint64) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	n, ok := r.names[alias]
	return n, ok
}

// Clear removes all mappings.
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.aliases = make(map[string]uint64)
	r.names = make(map[uint64]string)
}
