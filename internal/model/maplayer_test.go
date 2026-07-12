package model

import (
	"testing"

	"github.com/paulmach/orb"
)

func TestToMapLayerFeature_PointOnly(t *testing.T) {
	p := &POI{
		ID:       "way/1280560257",
		Name:     "Bill Baggs Cape Florida State Park",
		Location: orb.Point{-81.0, 28.0},
		Tags: map[string]string{
			"POI_NAME":           "Bathhouse Area 1",
			"POI_CLASSIFICATION": "Bathhouse",
		},
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
	if f.Properties["name"] != "Bathhouse Area 1 (Bathhouse)" {
		t.Fatalf("unexpected name: %v", f.Properties["name"])
	}
	if f.Properties["marker-color"] != "#7f8c8d" {
		t.Fatalf("expected bathhouse color, got %v", f.Properties["marker-color"])
	}
}

func TestToMapLayerFeature_BoundaryPolygon(t *testing.T) {
	p := &POI{
		ID:       "1",
		Name:     "Oleta River State Park",
		Source:   "geojson",
		Location: orb.Point{-80.13, 25.91},
		Tags: map[string]string{
			"source": "FL Park Boundaries",
		},
		Metadata: map[string]any{
			"geometry":      orb.Polygon{orb.Ring{{-80.14, 25.90}, {-80.12, 25.90}, {-80.12, 25.92}, {-80.14, 25.90}}},
			"geometry_type": "Polygon",
		},
	}
	f := ToMapLayerFeature(p)
	if f == nil {
		t.Fatal("expected feature")
	}
	if f.Geometry.GeoJSONType() != "Polygon" {
		t.Fatalf("expected Polygon, got %s", f.Geometry.GeoJSONType())
	}
	if f.Properties["fill-opacity"] != 0.25 {
		t.Fatalf("expected fill-opacity, got %v", f.Properties["fill-opacity"])
	}
}

func TestToMapLayerFeature_NumericID(t *testing.T) {
	p := &POI{ID: "42", Name: "Ramp", Location: orb.Point{-82.0, 29.0}, Tags: map[string]string{"RampName": "Sample Ramp"}}
	f := ToMapLayerFeature(p)
	if f == nil {
		t.Fatal("expected feature")
	}
	if f.ID != float64(42) {
		t.Fatalf("expected numeric id 42, got %v", f.ID)
	}
	if f.Properties["name"] != "Sample Ramp" {
		t.Fatalf("unexpected name: %v", f.Properties["name"])
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

func TestMapLayerStyleFor(t *testing.T) {
	cases := []struct {
		category string
		color    string
	}{
		{"Campground", "#27ae60"},
		{"Hand Launch Only", "#2980b9"},
		{"Hiking Trail", "#8e44ad"},
	}
	for _, tc := range cases {
		p := &POI{Location: orb.Point{-80, 25}, Category: tc.category}
		style := mapLayerStyleFor(p)
		if style.Color != tc.color {
			t.Fatalf("category %q: got color %s want %s", tc.category, style.Color, tc.color)
		}
	}
}
