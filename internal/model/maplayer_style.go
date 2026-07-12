package model

import (
	"strings"

	"github.com/paulmach/orb"
)

type mapLayerStyle struct {
	Color  string
	Size   string
	Symbol string
}

// styleRules match POI category text (case-insensitive substring).
var styleRules = []struct {
	match  []string
	style  mapLayerStyle
}{
	{[]string{"campground", "campsite", "cabin", "rv site", "primitive camping"}, mapLayerStyle{"#27ae60", "medium", "campsite"}},
	{[]string{"boat ramp", "ramp", "marina", "launch", "paddlecraft", "hand launch"}, mapLayerStyle{"#2980b9", "medium", "harbor"}},
	{[]string{"beach", "dune", "shore", "swim"}, mapLayerStyle{"#f39c12", "medium", "beach"}},
	{[]string{"trail", "hiking", "footpath", "nature trail"}, mapLayerStyle{"#8e44ad", "small", "park"}},
	{[]string{"restroom", "bathhouse", "toilet", "shower"}, mapLayerStyle{"#7f8c8d", "small", "toilets"}},
	{[]string{"picnic", "pavilion", "shelter", "grill"}, mapLayerStyle{"#e67e22", "small", "restaurant"}},
	{[]string{"campfire", "fire ring", "fire pit"}, mapLayerStyle{"#d35400", "small", "fire-station"}},
	{[]string{"entrance", "gate", "access point"}, mapLayerStyle{"#16a085", "small", "entrance"}},
	{[]string{"office", "ranger", "visitor center", "headquarters"}, mapLayerStyle{"#2c3e50", "medium", "town-hall"}},
	{[]string{"parking", "lot"}, mapLayerStyle{"#95a5a6", "small", "car"}},
	{[]string{"boundary", "state park", "park boundary"}, mapLayerStyle{"#27ae60", "large", "park"}},
	{[]string{"recreation", "playground", "ball field", "soccer", "baseball"}, mapLayerStyle{"#3498db", "medium", "pitch"}},
	{[]string{"fishing", "pier", "dock"}, mapLayerStyle{"#1abc9c", "medium", "wetland"}},
}

var boundaryFillStyle = mapLayerStyle{
	Color:  "#27ae60",
	Size:   "large",
	Symbol: "park",
}

func mapLayerStyleFor(p *POI) mapLayerStyle {
	if preserveBoundaryGeometry(p) {
		return boundaryFillStyle
	}
	text := strings.ToLower(mapLayerCategory(p) + " " + mapLayerName(p))
	for _, rule := range styleRules {
		for _, m := range rule.match {
			if strings.Contains(text, m) {
				return rule.style
			}
		}
	}
	return mapLayerStyle{mapLayerMarkerColor, mapLayerMarkerSize, "marker"}
}

func mapLayerCategory(p *POI) string {
	if p == nil {
		return ""
	}
	for _, key := range []string{
		"POI_CLASSIFICATION", "RampType", "CATEGORY", "type", "amenity",
		"tourism", "leisure", "HydrologicalType", "AccessType",
	} {
		if v := tagValue(p, key); v != "" {
			return v
		}
	}
	return p.Category
}

func tagValue(p *POI, key string) string {
	if p == nil {
		return ""
	}
	for k, v := range p.Tags {
		if strings.EqualFold(k, key) && strings.TrimSpace(v) != "" {
			return v
		}
	}
	if p.Metadata != nil {
		if v, ok := p.Metadata[key]; ok {
			if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
				return s
			}
		}
	}
	return ""
}

func preserveBoundaryGeometry(p *POI) bool {
	if p == nil {
		return false
	}
	src := strings.ToLower(layerSource(p))
	if strings.Contains(src, "boundar") {
		return polygonGeometry(p) != nil
	}
	return false
}

func layerSource(p *POI) string {
	if src := tagValue(p, "source"); src != "" {
		return src
	}
	return p.Source
}

func polygonGeometry(p *POI) orb.Geometry {
	if p == nil || p.Metadata == nil {
		return nil
	}
	g, ok := p.Metadata["geometry"].(orb.Geometry)
	if !ok || g == nil {
		return nil
	}
	switch g.(type) {
	case orb.Polygon, orb.MultiPolygon:
		return g
	default:
		return nil
	}
}
