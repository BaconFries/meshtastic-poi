package osm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"

	"github.com/BaconFries/meshtastic-poi/internal/downloader"
	"github.com/BaconFries/meshtastic-poi/internal/model"
	"github.com/BaconFries/meshtastic-poi/internal/providers"
)

const defaultOverpassURL = "https://overpass-api.de/api/interpreter"

// Provider downloads features from the OpenStreetMap Overpass API.
type Provider struct {
	cfg    providers.SourceConfig
	client *downloader.Client
}

type overpassResponse struct {
	Elements []overpassElement `json:"elements"`
}

type overpassElement struct {
	Type     string            `json:"type"`
	ID       int64             `json:"id"`
	Lat      float64           `json:"lat"`
	Lon      float64           `json:"lon"`
	Tags     map[string]string `json:"tags"`
	Geometry []overpassCoord   `json:"geometry"`
	Center   *overpassCoord    `json:"center"`
	Bounds   *overpassBounds   `json:"bounds"`
}

type overpassCoord struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type overpassBounds struct {
	MinLat float64 `json:"minlat"`
	MinLon float64 `json:"minlon"`
	MaxLat float64 `json:"maxlat"`
	MaxLon float64 `json:"maxlon"`
}

// New creates an Overpass API provider.
func New(cfg providers.SourceConfig, deps providers.Dependencies) (providers.Provider, error) {
	if cfg.URL == "" {
		cfg.URL = defaultOverpassURL
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
	return "osm"
}

func (p *Provider) Metadata(ctx context.Context) (*providers.DatasetInfo, error) {
	pois, err := p.Fetch(ctx)
	if err != nil {
		return nil, err
	}
	return &providers.DatasetInfo{
		Name:     p.Name(),
		Type:     "osm",
		URL:      p.cfg.URL,
		POICount: len(pois),
		Extra: map[string]any{
			"query": p.overpassQuery(),
		},
	}, nil
}

func (p *Provider) Fetch(ctx context.Context) ([]*model.POI, error) {
	query := p.overpassQuery()
	if query == "" {
		return nil, fmt.Errorf("osm provider requires params.query or bbox with params.tags")
	}

	form := url.Values{"data": {query}}
	body, err := p.client.Post(ctx, p.cfg.URL, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("overpass query: %w", err)
	}

	var resp overpassResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse overpass response: %w", err)
	}

	fc := &geojson.FeatureCollection{
		Type:     "FeatureCollection",
		Features: make([]*geojson.Feature, 0, len(resp.Elements)),
	}
	for _, el := range resp.Elements {
		f, ok := elementToFeature(el)
		if !ok {
			continue
		}
		fc.Features = append(fc.Features, f)
	}
	return model.FromFeatureCollection(fc, p.Name()), nil
}

func (p *Provider) overpassQuery() string {
	if q := strings.TrimSpace(p.cfg.Params["query"]); q != "" {
		return q
	}
	if q := strings.TrimSpace(p.cfg.Where); q != "" {
		return q
	}
	return p.buildBBoxQuery()
}

func (p *Provider) buildBBoxQuery() string {
	if len(p.cfg.BBox) != 4 {
		return ""
	}
	minLon, minLat, maxLon, maxLat := p.cfg.BBox[0], p.cfg.BBox[1], p.cfg.BBox[2], p.cfg.BBox[3]
	filter := p.tagFilter()
	if filter == "" {
		filter = `["amenity"]`
	}
	timeout := p.cfg.Params["timeout"]
	if timeout == "" {
		timeout = "90"
	}
	// Overpass bbox order: south, west, north, east
	bbox := fmt.Sprintf("(%f,%f,%f,%f)", minLat, minLon, maxLat, maxLon)
	return fmt.Sprintf(`[out:json][timeout:%s];
(
  node%s%s;
  way%s%s;
  relation%s%s;
);
out center geom;`, timeout, filter, bbox, filter, bbox, filter, bbox)
}

func (p *Provider) tagFilter() string {
	tags := strings.TrimSpace(p.cfg.Params["tags"])
	if tags == "" {
		return ""
	}
	parts := strings.Split(tags, ",")
	filters := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		k, v, ok := strings.Cut(part, "=")
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		if ok {
			filters = append(filters, fmt.Sprintf(`["%s"="%s"]`, k, strings.TrimSpace(v)))
		} else {
			filters = append(filters, fmt.Sprintf(`["%s"]`, k))
		}
	}
	return strings.Join(filters, "")
}

func elementToFeature(el overpassElement) (*geojson.Feature, bool) {
	props := make(geojson.Properties)
	for k, v := range el.Tags {
		props[k] = v
	}
	props["osm_type"] = el.Type
	props["osm_id"] = el.ID

	geom, ok := geometryFromElement(el)
	if !ok {
		return nil, false
	}

	f := geojson.NewFeature(geom)
	f.Properties = props
	f.ID = fmt.Sprintf("%s/%d", el.Type, el.ID)
	return f, true
}

func geometryFromElement(el overpassElement) (orb.Geometry, bool) {
	switch el.Type {
	case "node":
		if el.Lat == 0 && el.Lon == 0 {
			return nil, false
		}
		return orb.Point{el.Lon, el.Lat}, true
	case "way", "relation":
		if len(el.Geometry) >= 3 {
			ring := coordsToRing(el.Geometry)
			if ring[0] == ring[len(ring)-1] {
				return orb.Polygon{ring}, true
			}
			return ring, true
		}
		if el.Center != nil {
			return orb.Point{el.Center.Lon, el.Center.Lat}, true
		}
	}
	return nil, false
}

func coordsToRing(coords []overpassCoord) orb.Ring {
	ring := make(orb.Ring, len(coords))
	for i, c := range coords {
		ring[i] = orb.Point{c.Lon, c.Lat}
	}
	return ring
}

// Parse is exported for tests and converts raw Overpass JSON to GeoJSON.
func Parse(data []byte) (*geojson.FeatureCollection, error) {
	var resp overpassResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	fc := &geojson.FeatureCollection{Type: "FeatureCollection", Features: make([]*geojson.Feature, 0)}
	for _, el := range resp.Elements {
		f, ok := elementToFeature(el)
		if ok {
			fc.Features = append(fc.Features, f)
		}
	}
	return fc, nil
}
