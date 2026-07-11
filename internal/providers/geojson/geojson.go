package geojsonprovider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/paulmach/orb/geojson"

	"github.com/BaconFries/meshtastic-poi/internal/downloader"
	"github.com/BaconFries/meshtastic-poi/internal/model"
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

func (p *Provider) Metadata(ctx context.Context) (*providers.DatasetInfo, error) {
	pois, err := p.Fetch(ctx)
	if err != nil {
		return nil, err
	}
	return &providers.DatasetInfo{
		Name:     p.Name(),
		Type:     "geojson",
		URL:      p.cfg.URL,
		POICount: len(pois),
	}, nil
}

func (p *Provider) Fetch(ctx context.Context) ([]*model.POI, error) {
	body, err := source.Read(ctx, p.client, p.cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("download geojson: %w", err)
	}
	var fc geojson.FeatureCollection
	if err := json.Unmarshal(body, &fc); err != nil {
		return nil, fmt.Errorf("parse geojson: %w", err)
	}
	return model.FromFeatureCollection(&fc, p.Name()), nil
}
