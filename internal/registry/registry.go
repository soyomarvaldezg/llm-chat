package registry

import (
	"fmt"
	"sort"
	"sync"

	"github.com/soyomarvaldezg/llm-chat/internal/providers"
)

// Registry manages all available providers
type Registry struct {
	mu        sync.RWMutex
	providers map[string]providers.Provider
	metadata  map[string]providers.Metadata
}

// New creates a new provider registry
func New() *Registry {
	return &Registry{
		providers: make(map[string]providers.Provider),
		metadata:  make(map[string]providers.Metadata),
	}
}

// Register adds a provider to the registry
func (r *Registry) Register(provider providers.Provider, metadata providers.Metadata) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := provider.Name()
	if _, exists := r.providers[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}

	r.providers[name] = provider
	r.metadata[name] = metadata
	return nil
}

// Get retrieves a provider by name
func (r *Registry) Get(name string) (providers.Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, exists := r.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}

	return provider, nil
}

// GetMetadata retrieves metadata for a provider
func (r *Registry) GetMetadata(name string) (providers.Metadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metadata, exists := r.metadata[name]
	if !exists {
		return providers.Metadata{}, fmt.Errorf("metadata for provider %s not found", name)
	}

	return metadata, nil
}

// List returns all registered provider names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}

// ListAvailable returns only available (configured) providers
func (r *Registry) ListAvailable() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	available := make([]string, 0)
	for name, provider := range r.providers {
		if provider.IsAvailable() {
			available = append(available, name)
		}
	}

	sort.Strings(available)
	return available
}

// GetAll returns all providers with their metadata
func (r *Registry) GetAll() map[string]ProviderInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]ProviderInfo)
	for name, provider := range r.providers {
		result[name] = ProviderInfo{
			Provider:  provider,
			Metadata:  r.metadata[name],
			Available: provider.IsAvailable(),
		}
	}

	return result
}

// ProviderInfo combines provider, metadata, and availability
type ProviderInfo struct {
	Provider  providers.Provider
	Metadata  providers.Metadata
	Available bool
}

// Global registry instance
var defaultRegistry = New()

// Default returns the default global registry
func Default() *Registry {
	return defaultRegistry
}
