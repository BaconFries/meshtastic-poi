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

// ToMapLayerFeature converts a POI into a GeoJSON feature styled for Meshtastic map overlays.
func ToMapLayerFeature(p *POI) *geojson.Feature {
	if p == nil || !p.Valid() {
		return nil
	}

	style := mapLayerStyleFor(p)
	name := mapLayerName(p)
	category := mapLayerCategory(p)

	props := geojson.Properties{
		"name":           name,
		"marker-color":   style.Color,
		"marker-size":    style.Size,
		"marker-symbol":  style.Symbol,
		"category":       category,
	}
	if category != "" {
		props["type"] = category
	}
	if desc := mapLayerDescription(p); desc != "" {
		props["description"] = desc
	}

	var geom orb.Geometry = orb.Point{p.Location[0], p.Location[1]}
	if g := polygonGeometry(p); g != nil && preserveBoundaryGeometry(p) {
		geom = g
		props["fill"] = style.Color
		props["fill-opacity"] = 0.25
		props["stroke"] = style.Color
		props["stroke-width"] = 2
	}

	f := geojson.NewFeature(geom)
	f.Properties = props
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
	poiName := tagValue(p, "POI_NAME")
	classification := tagValue(p, "POI_CLASSIFICATION")
	siteName := firstNonEmpty(tagValue(p, "SITE_NAME"), tagValue(p, "PARK_NAME"), tagValue(p, "FACILITY_NAME"))
	rampName := tagValue(p, "RampName")

	if rampName != "" {
		return rampName
	}
	if poiName != "" {
		if classification != "" && !strings.EqualFold(poiName, classification) {
			return fmt.Sprintf("%s (%s)", poiName, classification)
		}
		return poiName
	}
	if classification != "" {
		if siteName != "" {
			return fmt.Sprintf("%s — %s", siteName, classification)
		}
		return classification
	}
	if p.Name != "" && (siteName == "" || !strings.EqualFold(p.Name, siteName)) {
		return p.Name
	}
	if siteName != "" {
		return siteName
	}
	if p.Category != "" {
		return p.Category
	}
	for _, key := range []string{"amenity", "tourism", "leisure", "name"} {
		if v := tagValue(p, key); v != "" {
			return strings.ReplaceAll(v, "_", " ")
		}
	}
	return fmt.Sprintf("POI %.5f,%.5f", p.Location[1], p.Location[0])
}

func mapLayerDescription(p *POI) string {
	parts := make([]string, 0, 4)
	for _, key := range []string{"SITE_NAME", "PARK_NAME", "COUNTY", "County", "CATEGORY", "POI_CLASSIFICATION"} {
		if v := tagValue(p, key); v != "" {
			parts = append(parts, v)
		}
	}
	if p.Description != "" {
		parts = append(parts, p.Description)
	}
	return strings.Join(uniqueStrings(parts), " · ")
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func uniqueStrings(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		key := strings.ToLower(s)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, s)
	}
	return out
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
