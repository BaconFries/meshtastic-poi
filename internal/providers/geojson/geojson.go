package geojsonprovider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/paulmach/orb/geojson"

	"github.com/BaconFries/meshtastic-poi/internal/downloader"
	"github.com/BaconFries/meshtastic-poi/internal/providers"
	"github.com/BaconFries/meshtastic-poi/internal/providers/source"
)

// Provider downloads GeoJSON from a URL.
type Provider struct {
	cfg    providers.SourceConfig
	client *downloader.Client
}

// New creates a GeoJSON URL provider.
func New(cfg providers.SourceConfig, deps providers.Dependencies) (providers.Provider, error) {
	return &Provider{
		cfg:    cfg,
		client: downloader.New(deps.CacheDir),
	}, nil
}

func (p *Provider) Name() string {
	if p.cfg.Name != "" {
		return p.cfg.Name
	}
	return "geojson"
}

func (p *Provider) Metadata(ctx context.Context) (*providers.Metadata, error) {
	fc, err := p.Download(ctx)
	if err != nil {
		return nil, err
	}
	geomType := ""
	if len(fc.Features) > 0 && fc.Features[0].Geometry != nil {
		geomType = fc.Features[0].Geometry.GeoJSONType()
	}
	return &providers.Metadata{
		Name:         p.Name(),
		Type:         "geojson",
		URL:          p.cfg.URL,
		FeatureCount: len(fc.Features),
		GeometryType: geomType,
	}, nil
}

func (p *Provider) Download(ctx context.Context) (*geojson.FeatureCollection, error) {
	body, err := source.Read(ctx, p.client, p.cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("download geojson: %w", err)
	}
	var fc geojson.FeatureCollection
	if err := json.Unmarshal(body, &fc); err != nil {
		return nil, fmt.Errorf("parse geojson: %w", err)
	}
	return &fc, nil
}
