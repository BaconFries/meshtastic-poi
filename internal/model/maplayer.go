package model

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
)

const (
	mapLayerMarkerColor = "#c0392b"
	mapLayerMarkerSize  = "medium"
)

// ToMapLayerFeature converts a POI into a Point-only GeoJSON feature styled for Meshtastic map overlays.
func ToMapLayerFeature(p *POI) *geojson.Feature {
	if p == nil || !p.Valid() {
		return nil
	}
	props := geojson.Properties{
		"name":          mapLayerName(p),
		"marker-color":  mapLayerMarkerColor,
		"marker-size":   mapLayerMarkerSize,
	}
	if p.Category != "" {
		props["type"] = p.Category
	}
	if p.Description != "" {
		props["description"] = p.Description
	}
	f := geojson.NewFeature(orb.Point{p.Location[0], p.Location[1]})
	f.Properties = props
	// Meshtastic-Apple decodes feature id as Int? only; string ids (e.g. "1", "way_123")
	// pass loose import validation but fail JSONDecoder when rendering overlays.
	if id, ok := mapLayerFeatureID(p.ID); ok {
		f.ID = id
	}
	return f
}

// ToMapLayerFeatureCollection converts POIs to a Meshtastic-friendly GeoJSON FeatureCollection.
func ToMapLayerFeatureCollection(pois []*POI) *geojson.FeatureCollection {
	features := make([]*geojson.Feature, 0, len(pois))
	for _, p := range pois {
		if f := ToMapLayerFeature(p); f != nil {
			features = append(features, f)
		}
	}
	return &geojson.FeatureCollection{Type: "FeatureCollection", Features: features}
}

func mapLayerName(p *POI) string {
	if p.Name != "" {
		return p.Name
	}
	if p.Category != "" {
		return p.Category
	}
	for _, key := range []string{"amenity", "tourism", "leisure", "name"} {
		if v := p.Tags[key]; v != "" {
			return strings.ReplaceAll(v, "_", " ")
		}
	}
	return fmt.Sprintf("POI %.5f,%.5f", p.Location[1], p.Location[0])
}

// mapLayerFeatureID returns a numeric GeoJSON feature id when safe for Meshtastic-Apple.
func mapLayerFeatureID(id string) (float64, bool) {
	id = strings.TrimSpace(id)
	if id == "" {
		return 0, false
	}
	n, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, false
	}
	return float64(n), true
}
