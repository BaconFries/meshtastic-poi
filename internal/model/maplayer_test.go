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
	if f.ID != nil {
		t.Fatalf("expected no id for OSM-style ids, got %v", f.ID)
	}
	if f.Properties["marker-color"] != mapLayerMarkerColor {
		t.Fatalf("expected marker-color, got %v", f.Properties["marker-color"])
	}
}

func TestToMapLayerFeature_NumericID(t *testing.T) {
	p := &POI{ID: "42", Name: "Ramp", Location: orb.Point{-82.0, 29.0}}
	f := ToMapLayerFeature(p)
	if f == nil {
		t.Fatal("expected feature")
	}
	if f.ID != float64(42) {
		t.Fatalf("expected numeric id 42, got %v", f.ID)
	}
}

func TestMapLayerFeatureID(t *testing.T) {
	if _, ok := mapLayerFeatureID("node_10010311085"); ok {
		t.Fatal("expected non-numeric id to be omitted")
	}
	if n, ok := mapLayerFeatureID("1001"); !ok || n != 1001 {
		t.Fatalf("got %v %v", n, ok)
	}
}
