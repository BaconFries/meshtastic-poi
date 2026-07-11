// Package exporters writes canonical POI datasets to external formats.
package exporters

import (
	"context"
	"fmt"
	"io"

	"github.com/BaconFries/meshtastic-poi/internal/model"
)

// Exporter serializes POI datasets to a target format.
type Exporter interface {
	Name() string
	Export(ctx context.Context, pois []*model.POI, w io.Writer) error
}

// Registry maps exporter names to implementations.
type Registry struct {
	exporters map[string]Exporter
}

// NewRegistry returns an empty exporter registry.
func NewRegistry() *Registry {
	return &Registry{exporters: make(map[string]Exporter)}
}

// Register adds an exporter.
func (r *Registry) Register(name string, e Exporter) {
	r.exporters[name] = e
}

// Get returns an exporter by name.
func (r *Registry) Get(name string) (Exporter, error) {
	e, ok := r.exporters[name]
	if !ok {
		return nil, fmt.Errorf("unknown exporter %q", name)
	}
	return e, nil
}

// Types returns registered exporter names.
func (r *Registry) Types() []string {
	types := make([]string, 0, len(r.exporters))
	for t := range r.exporters {
		types = append(types, t)
	}
	return types
}
