package model_test

import (
	"testing"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"

	"github.com/BaconFries/meshtastic-poi/internal/model"
)

func TestFromFeatureRoundTrip(t *testing.T) {
	f := geojson.NewFeature(orb.Point{-81.3792, 28.5383})
	f.Properties = geojson.Properties{
		"name":        "Orlando",
		"type":        "park",
		"description": "Test",
	}
	f.ID = "1"

	poi := model.FromFeature(f, "test")
	if poi.Name != "Orlando" || poi.Category != "park" {
		t.Fatalf("unexpected poi fields: %+v", poi)
	}

	back := model.ToFeature(poi)
	if back.Properties["name"] != "Orlando" {
		t.Fatalf("round trip failed: %+v", back.Properties)
	}
}

func TestMerge(t *testing.T) {
	a := []*model.POI{{ID: "1"}, {ID: "2"}}
	b := []*model.POI{{ID: "3"}}
	merged := model.Merge(a, b)
	if len(merged) != 3 {
		t.Fatalf("expected 3, got %d", len(merged))
	}
}
