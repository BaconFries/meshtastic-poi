package providers

import (
	"context"
	"fmt"

	"github.com/paulmach/orb/geojson"
)

// Metadata describes a provider data source.
type Metadata struct {
	Name           string         `json:"name"`
	Type           string         `json:"type"`
	URL            string         `json:"url,omitempty"`
	FeatureCount   int            `json:"feature_count,omitempty"`
	MaxRecordCount int            `json:"max_record_count,omitempty"`
	GeometryType   string         `json:"geometry_type,omitempty"`
	Fields         []string       `json:"fields,omitempty"`
	Extra          map[string]any `json:"extra,omitempty"`
}

// SourceConfig describes a configured data source.
type SourceConfig struct {
	Name   string            `mapstructure:"name" validate:"required"`
	Type   string            `mapstructure:"type" validate:"required"`
	URL    string            `mapstructure:"url" validate:"required"`
	Where  string            `mapstructure:"where"`
	BBox   []float64         `mapstructure:"bbox"`
	Params map[string]string `mapstructure:"params"`
}

// Provider downloads and describes geospatial data sources.
type Provider interface {
	Name() string
	Download(ctx context.Context) (*geojson.FeatureCollection, error)
	Metadata(ctx context.Context) (*Metadata, error)
}

// Factory creates a provider from source configuration.
type Factory func(cfg SourceConfig, deps Dependencies) (Provider, error)

// Dependencies shared across providers.
type Dependencies struct {
	CacheDir string
}

// Registry maps provider type names to factories.
type Registry struct {
	factories map[string]Factory
}

// NewRegistry returns an empty provider registry.
func NewRegistry() *Registry {
	return &Registry{factories: make(map[string]Factory)}
}

// Register adds a provider factory.
func (r *Registry) Register(typeName string, factory Factory) {
	r.factories[typeName] = factory
}

// Create instantiates a provider for the given source config.
func (r *Registry) Create(cfg SourceConfig, deps Dependencies) (Provider, error) {
	factory, ok := r.factories[cfg.Type]
	if !ok {
		return nil, fmt.Errorf("unknown provider type %q", cfg.Type)
	}
	return factory(cfg, deps)
}

// Types returns registered provider type names.
func (r *Registry) Types() []string {
	types := make([]string, 0, len(r.factories))
	for t := range r.factories {
		types = append(types, t)
	}
	return types
}
