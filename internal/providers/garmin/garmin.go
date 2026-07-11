package garmin

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

// Provider reads Garmin Custom POI CSV exports.
type Provider struct {
	cfg    providers.SourceConfig
	client *downloader.Client
}

// New creates a Garmin POI CSV provider.
func New(cfg providers.SourceConfig, deps providers.Dependencies) (providers.Provider, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("garmin provider requires url")
	}
	return &Provider{
		cfg:    cfg,
		client: downloader.New(deps.CacheDir),
	}, nil
}

func (p *Provider) Name() string {
	if p.cfg.Name != "" {
		return p.cfg.Name
	}
	return "garmin"
}

func (p *Provider) Metadata(ctx context.Context) (*providers.DatasetInfo, error) {
	pois, err := p.Fetch(ctx)
	if err != nil {
		return nil, err
	}
	return &providers.DatasetInfo{
		Name:     p.Name(),
		Type:     "garmin",
		URL:      p.cfg.URL,
		POICount: len(pois),
		Extra: map[string]any{
			"format": "garmin_csv",
		},
	}, nil
}

func (p *Provider) Fetch(ctx context.Context) ([]*model.POI, error) {
	body, err := source.Read(ctx, p.client, p.cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("read garmin poi: %w", err)
	}
	fc, err := ParseCSV(body, p.columnMap())
	if err != nil {
		return nil, err
	}
	return model.FromFeatureCollection(fc, p.Name()), nil
}

func (p *Provider) columnMap() map[string]string {
	m := map[string]string{
		"lat":  "latitude",
		"lon":  "longitude",
		"name": "name",
		"desc": "description",
		"type": "category",
	}
	for k, v := range p.cfg.Params {
		switch strings.ToLower(k) {
		case "lat_column", "latitude_column":
			m["lat"] = v
		case "lon_column", "longitude_column":
			m["lon"] = v
		case "name_column":
			m["name"] = v
		case "desc_column", "description_column":
			m["desc"] = v
		case "type_column", "category_column":
			m["type"] = v
		}
	}
	return m
}

// ParseCSV converts Garmin POI CSV bytes to GeoJSON.
func ParseCSV(data []byte, columns map[string]string) (*geojson.FeatureCollection, error) {
	reader := csv.NewReader(strings.NewReader(string(data)))
	reader.TrimLeadingSpace = true
	reader.LazyQuotes = true

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read garmin csv header: %w", err)
	}

	colIndex := make(map[string]int)
	for i, h := range header {
		colIndex[normalizeCol(h)] = i
	}

	latIdx := findColumn(colIndex, columns["lat"], "latitude", "lat", "y")
	lonIdx := findColumn(colIndex, columns["lon"], "longitude", "lon", "long", "x")
	if latIdx < 0 || lonIdx < 0 {
		return nil, fmt.Errorf("garmin csv missing latitude/longitude columns")
	}

	nameIdx := findColumn(colIndex, columns["name"], "name", "poi name", "title")
	descIdx := findColumn(colIndex, columns["desc"], "description", "desc", "comment", "notes")
	typeIdx := findColumn(colIndex, columns["type"], "category", "type", "symbol", "class")

	fc := &geojson.FeatureCollection{Type: "FeatureCollection", Features: make([]*geojson.Feature, 0)}
	rowNum := 1
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read garmin csv row: %w", err)
		}
		rowNum++

		lat, err1 := parseCoord(row, latIdx)
		lon, err2 := parseCoord(row, lonIdx)
		if err1 != nil || err2 != nil {
			continue
		}

		props := geojson.Properties{"source": "garmin"}
		for i, h := range header {
			if i == latIdx || i == lonIdx {
				continue
			}
			if strings.TrimSpace(row[i]) != "" {
				props[h] = row[i]
			}
		}
		if nameIdx >= 0 {
			props["name"] = row[nameIdx]
		}
		if descIdx >= 0 {
			props["description"] = row[descIdx]
		}
		if typeIdx >= 0 {
			props["category"] = row[typeIdx]
		}

		feature := geojson.NewFeature(orb.Point{lon, lat})
		feature.Properties = props
		feature.ID = rowNum
		fc.Features = append(fc.Features, feature)
	}
	return fc, nil
}

func normalizeCol(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func findColumn(index map[string]int, preferred string, fallbacks ...string) int {
	if preferred != "" {
		if i, ok := index[normalizeCol(preferred)]; ok {
			return i
		}
	}
	for _, name := range fallbacks {
		if i, ok := index[normalizeCol(name)]; ok {
			return i
		}
	}
	return -1
}

func parseCoord(row []string, idx int) (float64, error) {
	if idx < 0 || idx >= len(row) {
		return 0, fmt.Errorf("missing column")
	}
	return strconv.ParseFloat(strings.TrimSpace(row[idx]), 64)
}
