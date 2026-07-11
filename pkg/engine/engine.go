// Package engine provides the public POI management API for embedding in other applications.
package engine

import (
	"context"

	"github.com/BaconFries/meshtastic-poi/internal/exporters"
	"github.com/BaconFries/meshtastic-poi/internal/model"
	"github.com/BaconFries/meshtastic-poi/internal/pipeline"
	"github.com/BaconFries/meshtastic-poi/internal/providers"
	"github.com/BaconFries/meshtastic-poi/internal/providers/register"
	"github.com/BaconFries/meshtastic-poi/internal/spatial"
)

// Fetch downloads POIs from a configured provider source.
func Fetch(ctx context.Context, src providers.SourceConfig, deps providers.Dependencies) ([]*model.POI, error) {
	p, err := register.DefaultRegistry().Create(src, deps)
	if err != nil {
		return nil, err
	}
	return p.Fetch(ctx)
}

// Process runs the default optimization pipeline on POIs.
func Process(ctx context.Context, pois []*model.POI, opts pipeline.Options) ([]*model.POI, error) {
	return pipeline.Run(ctx, pois, pipeline.Default(opts))
}

// Validate returns a validation report for POIs.
func Validate(pois []*model.POI) pipeline.Report {
	return pipeline.ValidateReport(pois)
}

// Export writes POIs using the named exporter format.
func Export(path, format string, pois []*model.POI) error {
	return exporters.WriteFile(path, format, pois)
}

// Index builds a spatial index over POIs.
func Index(pois []*model.POI) *spatial.POIIndex {
	return spatial.NewPOIIndex(pois)
}

// LoadGeoJSON imports POIs from a GeoJSON file.
func LoadGeoJSON(path string) ([]*model.POI, error) {
	return exporters.ReadGeoJSONFile(path)
}

// SaveGeoJSON exports POIs to a GeoJSON file.
func SaveGeoJSON(path string, pois []*model.POI) error {
	return exporters.WriteGeoJSONFile(path, pois)
}

// Merge combines multiple POI slices.
func Merge(slices ...[]*model.POI) []*model.POI {
	return model.Merge(slices...)
}
