package model

import (
	"fmt"
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
	if id := sanitizeFeatureID(p.ID); id != "" {
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

func sanitizeFeatureID(id string) string {
	id = strings.TrimSpace(id)
	if id == "" {
		return ""
	}
	id = strings.ReplaceAll(id, "/", "_")
	id = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_', r == '-':
			return r
		default:
			return '_'
		}
	}, id)
	return strings.Trim(id, "_")
}
