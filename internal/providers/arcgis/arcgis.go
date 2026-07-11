package arcgis

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"

	"github.com/BaconFries/meshtastic-poi/internal/downloader"
	"github.com/BaconFries/meshtastic-poi/internal/model"
	"github.com/BaconFries/meshtastic-poi/internal/providers"
)

const defaultWhere = "1=1"

// Provider downloads features from an ArcGIS FeatureServer layer.
type Provider struct {
	cfg    providers.SourceConfig
	client *downloader.Client
	meta   *layerInfo
}

type layerInfo struct {
	Name           string
	MaxRecordCount int
	GeometryType   string
	Fields         []field
	FeatureCount   int
}

type layerResponse struct {
	Name           string  `json:"name"`
	MaxRecordCount int     `json:"maxRecordCount"`
	GeometryType   string  `json:"geometryType"`
	Fields         []field `json:"fields"`
}

type field struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type queryResponse struct {
	Features              []arcgisFeature `json:"features"`
	ExceededTransferLimit bool            `json:"exceededTransferLimit"`
	Error                 *arcgisError    `json:"error"`
}

type geojsonQueryResponse struct {
	Type                  string             `json:"type"`
	Features              []*geojson.Feature `json:"features"`
	ExceededTransferLimit bool               `json:"exceededTransferLimit"`
	Error                 *arcgisError       `json:"error"`
}

type arcgisError struct {
	Code        int    `json:"code"`
	Message     string `json:"message"`
	Description string `json:"description"`
}

type arcgisFeature struct {
	Attributes map[string]any  `json:"attributes"`
	Geometry   json.RawMessage `json:"geometry"`
}

// New creates an ArcGIS FeatureServer provider.
func New(cfg providers.SourceConfig, deps providers.Dependencies) (providers.Provider, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("arcgis provider requires url")
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
	return "arcgis"
}

func (p *Provider) layerURL() string {
	u := strings.TrimRight(p.cfg.URL, "/")
	if !strings.HasSuffix(u, "/query") {
		return u + "/query"
	}
	return u
}

func (p *Provider) infoURL() string {
	u := strings.TrimRight(p.cfg.URL, "/")
	u = strings.TrimSuffix(u, "/query")
	return u + "?f=json"
}

func (p *Provider) Metadata(ctx context.Context) (*providers.DatasetInfo, error) {
	info, err := p.fetchLayerInfo(ctx)
	if err != nil {
		return nil, err
	}
	fields := make([]string, len(info.Fields))
	for i, f := range info.Fields {
		fields[i] = f.Name
	}
	return &providers.DatasetInfo{
		Name:           p.Name(),
		Type:           "arcgis",
		URL:            p.cfg.URL,
		POICount:       info.FeatureCount,
		MaxRecordCount: info.MaxRecordCount,
		Fields:         fields,
		Extra: map[string]any{
			"geometry_type": info.GeometryType,
		},
	}, nil
}

func (p *Provider) Fetch(ctx context.Context) ([]*model.POI, error) {
	fc, err := p.fetchFeatures(ctx)
	if err != nil {
		return nil, err
	}
	return model.FromFeatureCollection(fc, p.Name()), nil
}

func (p *Provider) fetchFeatures(ctx context.Context) (*geojson.FeatureCollection, error) {
	info, err := p.fetchLayerInfo(ctx)
	if err != nil {
		return nil, err
	}
	p.meta = info

	where := p.cfg.Where
	if where == "" {
		where = defaultWhere
	}

	pageSize := info.MaxRecordCount
	if pageSize <= 0 {
		pageSize = 1000
	}

	fc := &geojson.FeatureCollection{Features: make([]*geojson.Feature, 0)}
	offset := 0

	for {
		features, exceeded, err := p.queryPage(ctx, where, offset, pageSize)
		if err != nil {
			return nil, err
		}
		fc.Features = append(fc.Features, features...)
		if !exceeded || len(features) == 0 {
			break
		}
		offset += len(features)
	}

	return fc, nil
}

func (p *Provider) fetchLayerInfo(ctx context.Context) (*layerInfo, error) {
	body, err := p.client.GetJSON(ctx, p.infoURL())
	if err != nil {
		return nil, fmt.Errorf("fetch layer info: %w", err)
	}
	var resp layerResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse layer info: %w", err)
	}
	return &layerInfo{
		Name:           resp.Name,
		MaxRecordCount: resp.MaxRecordCount,
		GeometryType:   resp.GeometryType,
		Fields:         resp.Fields,
	}, nil
}

func (p *Provider) queryPage(ctx context.Context, where string, offset, limit int) ([]*geojson.Feature, bool, error) {
	params := url.Values{}
	params.Set("f", "geojson")
	params.Set("where", where)
	params.Set("outFields", "*")
	params.Set("returnGeometry", "true")
	params.Set("resultOffset", strconv.Itoa(offset))
	params.Set("resultRecordCount", strconv.Itoa(limit))

	if len(p.cfg.BBox) == 4 {
		params.Set("geometry", fmt.Sprintf("%f,%f,%f,%f", p.cfg.BBox[0], p.cfg.BBox[1], p.cfg.BBox[2], p.cfg.BBox[3]))
		params.Set("geometryType", "esriGeometryEnvelope")
		params.Set("spatialRel", "esriSpatialRelIntersects")
		params.Set("inSR", "4326")
	}

	for k, v := range p.cfg.Params {
		params.Set(k, v)
	}

	queryURL := p.layerURL() + "?" + params.Encode()
	body, err := p.client.GetJSON(ctx, queryURL)
	if err != nil {
		return nil, false, err
	}

	var gjResp geojsonQueryResponse
	if err := json.Unmarshal(body, &gjResp); err == nil && (len(gjResp.Features) > 0 || gjResp.Type == "FeatureCollection") {
		if gjResp.Error != nil {
			return nil, false, fmt.Errorf("arcgis error %d: %s", gjResp.Error.Code, gjResp.Error.Message)
		}
		exceeded := gjResp.ExceededTransferLimit
		if !exceeded && len(gjResp.Features) >= limit {
			exceeded = true
		}
		return gjResp.Features, exceeded, nil
	}

	var resp queryResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, false, fmt.Errorf("parse query response: %w", err)
	}

	if resp.Error != nil {
		return nil, false, fmt.Errorf("arcgis error %d: %s", resp.Error.Code, resp.Error.Message)
	}

	features := make([]*geojson.Feature, 0, len(resp.Features))
	for _, af := range resp.Features {
		f, err := convertFeature(af)
		if err != nil {
			continue
		}
		features = append(features, f)
	}
	return features, resp.ExceededTransferLimit, nil
}

func convertFeature(af arcgisFeature) (*geojson.Feature, error) {
	f := &geojson.Feature{
		Type:       "Feature",
		Properties: af.Attributes,
	}
	if len(af.Geometry) == 0 || string(af.Geometry) == "null" {
		return f, nil
	}

	var gj geojson.Geometry
	if err := json.Unmarshal(af.Geometry, &gj); err == nil && gj.Coordinates != nil {
		f.Geometry = gj.Geometry()
		return f, nil
	}

	var esri map[string]any
	if err := json.Unmarshal(af.Geometry, &esri); err != nil {
		return nil, err
	}
	geom, ok := esriToOrb(esri)
	if !ok {
		return f, nil
	}
	f.Geometry = geom
	return f, nil
}

func esriToOrb(esri map[string]any) (orb.Geometry, bool) {
	if x, ok := esri["x"].(float64); ok {
		if y, ok := esri["y"].(float64); ok {
			return orb.Point{x, y}, true
		}
	}
	if paths, ok := esri["paths"].([]any); ok && len(paths) > 0 {
		if line, ok := pathToLineString(paths[0]); ok {
			return line, true
		}
	}
	if rings, ok := esri["rings"].([]any); ok && len(rings) > 0 {
		if poly, ok := ringToPolygon(rings[0]); ok {
			return poly, true
		}
	}
	return nil, false
}

func pathToLineString(path any) (orb.LineString, bool) {
	pts, ok := path.([]any)
	if !ok || len(pts) == 0 {
		return nil, false
	}
	line := make(orb.LineString, 0, len(pts))
	for _, pt := range pts {
		coords, ok := pt.([]any)
		if !ok || len(coords) < 2 {
			continue
		}
		x, _ := coords[0].(float64)
		y, _ := coords[1].(float64)
		line = append(line, orb.Point{x, y})
	}
	return line, len(line) > 0
}

func ringToPolygon(ring any) (orb.Polygon, bool) {
	pts, ok := ring.([]any)
	if !ok || len(pts) == 0 {
		return nil, false
	}
	ringLS := make(orb.Ring, 0, len(pts))
	for _, pt := range pts {
		coords, ok := pt.([]any)
		if !ok || len(coords) < 2 {
			continue
		}
		x, _ := coords[0].(float64)
		y, _ := coords[1].(float64)
		ringLS = append(ringLS, orb.Point{x, y})
	}
	if len(ringLS) == 0 {
		return nil, false
	}
	return orb.Polygon{ringLS}, true
}

func init() {
	// registered via providers.RegisterAll
}
