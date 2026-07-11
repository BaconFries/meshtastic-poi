package kml

import (
	"context"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"

	"github.com/BaconFries/meshtastic-poi/internal/downloader"
	"github.com/BaconFries/meshtastic-poi/internal/providers"
	"github.com/BaconFries/meshtastic-poi/internal/providers/source"
)

// Provider reads KML files from URL or local path.
type Provider struct {
	cfg    providers.SourceConfig
	client *downloader.Client
}

type kmlDoc struct {
	XMLName    xml.Name       `xml:"kml"`
	Document   kmlContainer   `xml:"Document"`
	Folder     kmlContainer   `xml:"Folder"`
	Placemarks []kmlPlacemark `xml:"Placemark"`
}

type kmlContainer struct {
	Name       string         `xml:"name"`
	Placemarks []kmlPlacemark `xml:"Placemark"`
	Folders    []kmlContainer `xml:"Folder"`
}

type kmlPlacemark struct {
	Name        string      `xml:"name"`
	Description string      `xml:"description"`
	Point       *kmlPoint   `xml:"Point"`
	LineString  *kmlLine    `xml:"LineString"`
	Polygon     *kmlPolygon `xml:"Polygon"`
}

type kmlPoint struct {
	Coordinates string `xml:"coordinates"`
}

type kmlLine struct {
	Coordinates string `xml:"coordinates"`
}

type kmlPolygon struct {
	OuterBoundary kmlBoundary `xml:"outerBoundaryIs>LinearRing"`
}

type kmlBoundary struct {
	Coordinates string `xml:"coordinates"`
}

// New creates a KML provider.
func New(cfg providers.SourceConfig, deps providers.Dependencies) (providers.Provider, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("kml provider requires url")
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
	return "kml"
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
		Type:         "kml",
		URL:          p.cfg.URL,
		FeatureCount: len(fc.Features),
		GeometryType: geomType,
	}, nil
}

func (p *Provider) Download(ctx context.Context) (*geojson.FeatureCollection, error) {
	body, err := source.Read(ctx, p.client, p.cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("read kml: %w", err)
	}
	return Parse(body)
}

// Parse converts KML XML bytes to GeoJSON.
func Parse(data []byte) (*geojson.FeatureCollection, error) {
	var doc kmlDoc
	if err := xml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse kml: %w", err)
	}

	fc := &geojson.FeatureCollection{
		Type:     "FeatureCollection",
		Features: make([]*geojson.Feature, 0),
	}
	idx := 0
	collect := func(pm kmlPlacemark) {
		f, ok := placemarkToFeature(pm, idx)
		if ok {
			fc.Features = append(fc.Features, f)
			idx++
		}
	}
	for _, pm := range doc.Placemarks {
		collect(pm)
	}
	walkContainer(doc.Document, collect)
	walkContainer(doc.Folder, collect)
	return fc, nil
}

func walkContainer(c kmlContainer, fn func(kmlPlacemark)) {
	for _, pm := range c.Placemarks {
		fn(pm)
	}
	for _, folder := range c.Folders {
		walkContainer(folder, fn)
	}
}

func placemarkToFeature(pm kmlPlacemark, idx int) (*geojson.Feature, bool) {
	var geom orb.Geometry
	switch {
	case pm.Point != nil:
		pt, ok := parseFirstCoord(pm.Point.Coordinates)
		if !ok {
			return nil, false
		}
		geom = pt
	case pm.LineString != nil:
		line, ok := parseLine(pm.LineString.Coordinates)
		if !ok {
			return nil, false
		}
		geom = line
	case pm.Polygon != nil:
		ring, ok := parseRing(pm.Polygon.OuterBoundary.Coordinates)
		if !ok {
			return nil, false
		}
		geom = orb.Polygon{ring}
	default:
		return nil, false
	}

	f := geojson.NewFeature(geom)
	f.Properties = geojson.Properties{
		"name":        pm.Name,
		"description": pm.Description,
		"kml":         true,
	}
	f.ID = "placemark/" + strconv.Itoa(idx)
	return f, true
}

func parseFirstCoord(raw string) (orb.Point, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return orb.Point{}, false
	}
	parts := strings.Split(raw, ",")
	if len(parts) < 2 {
		return orb.Point{}, false
	}
	lon, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	lat, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err1 != nil || err2 != nil {
		return orb.Point{}, false
	}
	return orb.Point{lon, lat}, true
}

func parseLine(raw string) (orb.LineString, bool) {
	pts := parseCoordList(raw)
	if len(pts) < 2 {
		return nil, false
	}
	return pts, true
}

func parseRing(raw string) (orb.Ring, bool) {
	pts := parseCoordList(raw)
	if len(pts) < 3 {
		return nil, false
	}
	ring := orb.Ring(pts)
	if ring[0] != ring[len(ring)-1] {
		ring = append(ring, ring[0])
	}
	return ring, true
}

func parseCoordList(raw string) orb.LineString {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	tokens := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ' ' || r == '\n' || r == '\r' || r == '\t'
	})
	line := make(orb.LineString, 0, len(tokens))
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		if pt, ok := parseFirstCoord(token); ok {
			line = append(line, pt)
		}
	}
	return line
}
