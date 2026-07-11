package model

import (
	"testing"

	"github.com/paulmach/orb"
)

func TestToMapLayerFeature_PointOnly(t *testing.T) {
	p := &POI{
		ID:       "way/1280560257",
		Name:     "Test Station",
		Location: orb.Point{-81.0, 28.0},
		Metadata: map[string]any{
			"geometry": orb.Polygon{orb.Ring{{-81.1, 28.1}, {-81.0, 28.1}, {-81.0, 28.0}, {-81.1, 28.1}}},
		},
	}
	f := ToMapLayerFeature(p)
	if f == nil {
		t.Fatal("expected feature")
	}
	if f.Geometry.GeoJSONType() != "Point" {
		t.Fatalf("expected Point, got %s", f.Geometry.GeoJSONType())
	}
	if f.ID != "way_1280560257" {
		t.Fatalf("expected sanitized id, got %v", f.ID)
	}
	if f.Properties["marker-color"] != mapLayerMarkerColor {
		t.Fatalf("expected marker-color, got %v", f.Properties["marker-color"])
	}
}

func TestSanitizeFeatureID(t *testing.T) {
	if got := sanitizeFeatureID("node/10010311085"); got != "node_10010311085" {
		t.Fatalf("got %q", got)
	}
}
