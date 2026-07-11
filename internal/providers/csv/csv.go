package csvprovider

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"

	"github.com/BaconFries/meshtastic-poi/internal/downloader"
	"github.com/BaconFries/meshtastic-poi/internal/model"
	"github.com/BaconFries/meshtastic-poi/internal/providers"
	"github.com/BaconFries/meshtastic-poi/internal/providers/source"
)

// Provider downloads CSV with lat/lon columns and converts to GeoJSON points.
type Provider struct {
	cfg    providers.SourceConfig
	client *downloader.Client
}

// New creates a CSV provider.
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
	return "csv"
}

func (p *Provider) Metadata(ctx context.Context) (*providers.DatasetInfo, error) {
	pois, err := p.Fetch(ctx)
	if err != nil {
		return nil, err
	}
	return &providers.DatasetInfo{
		Name:     p.Name(),
		Type:     "csv",
		URL:      p.cfg.URL,
		POICount: len(pois),
	}, nil
}

func (p *Provider) Fetch(ctx context.Context) ([]*model.POI, error) {
	body, err := source.Read(ctx, p.client, p.cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("download csv: %w", err)
	}

	latCol := p.cfg.Params["lat_column"]
	if latCol == "" {
		latCol = "lat"
	}
	lonCol := p.cfg.Params["lon_column"]
	if lonCol == "" {
		lonCol = "lon"
	}

	reader := csv.NewReader(strings.NewReader(string(body)))
	reader.TrimLeadingSpace = true
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read csv header: %w", err)
	}

	colIndex := make(map[string]int)
	for i, h := range header {
		colIndex[strings.ToLower(strings.TrimSpace(h))] = i
	}

	latIdx, okLat := colIndex[strings.ToLower(latCol)]
	lonIdx, okLon := colIndex[strings.ToLower(lonCol)]
	if !okLat || !okLon {
		return nil, fmt.Errorf("csv missing lat/lon columns %q/%q", latCol, lonCol)
	}

	fc := &geojson.FeatureCollection{Features: make([]*geojson.Feature, 0)}
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read csv row: %w", err)
		}
		lat, err1 := strconv.ParseFloat(strings.TrimSpace(row[latIdx]), 64)
		lon, err2 := strconv.ParseFloat(strings.TrimSpace(row[lonIdx]), 64)
		if err1 != nil || err2 != nil {
			continue
		}

		props := make(map[string]any)
		for i, h := range header {
			if i == latIdx || i == lonIdx {
				continue
			}
			props[h] = row[i]
		}

		feature := geojson.NewFeature(orb.Point{lon, lat})
		feature.Properties = props
		fc.Features = append(fc.Features, feature)
	}
	return model.FromFeatureCollection(fc, p.Name()), nil
}
