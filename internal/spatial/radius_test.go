package spatial_test

import (
	"path/filepath"
	"testing"

	"github.com/BaconFries/meshtastic-poi/internal/output"
	"github.com/BaconFries/meshtastic-poi/internal/spatial"
)

func TestFilterRadius(t *testing.T) {
	fc, err := output.ReadGeoJSON(filepath.Join("..", "..", "testdata", "sample_pois.geojson"))
	if err != nil {
		t.Fatal(err)
	}
	clean := fc
	// Orlando lat/lon with 50km radius should include Orlando duplicate pair and exclude Miami
	result := spatial.FilterRadius(clean, 28.5383, -81.3792, 50000)
	if len(result.Features) < 2 {
		t.Fatalf("expected at least 2 features near Orlando, got %d", len(result.Features))
	}
	for _, f := range result.Features {
		name, _ := f.Properties["name"].(string)
		if name == "Miami Beach" {
			t.Fatal("Miami Beach should be outside 50km radius of Orlando")
		}
	}
}

func TestFilterBBox(t *testing.T) {
	fc, err := output.ReadGeoJSON(filepath.Join("..", "..", "testdata", "sample_pois.geojson"))
	if err != nil {
		t.Fatal(err)
	}
	// Bounding box around central Florida
	result := spatial.FilterBBox(fc, -82.5, 27.5, -81.0, 29.0)
	if len(result.Features) < 2 {
		t.Fatalf("expected features in bbox, got %d", len(result.Features))
	}
}

func TestFilterAttributes(t *testing.T) {
	fc, err := output.ReadGeoJSON(filepath.Join("..", "..", "testdata", "sample_pois.geojson"))
	if err != nil {
		t.Fatal(err)
	}
	result := spatial.FilterAttributes(fc, map[string]string{"county": "Orange"})
	if len(result.Features) != 2 {
		t.Fatalf("expected 2 Orange county features, got %d", len(result.Features))
	}
}

func TestHaversineMeters(t *testing.T) {
	// Orlando to Tampa ~125km
	d := spatial.HaversineMeters(28.5383, -81.3792, 27.9506, -82.4572)
	if d < 100000 || d > 150000 {
		t.Fatalf("unexpected distance Orlando-Tampa: %f", d)
	}
}

func TestIndexDuplicateDetection(t *testing.T) {
	fc, err := output.ReadGeoJSON(filepath.Join("..", "..", "testdata", "dedupe_test.geojson"))
	if err != nil {
		t.Fatal(err)
	}
	idx := spatial.NewIndex(fc)
	dupes := idx.FindDuplicateCoordinates(6)
	if len(dupes) != 1 {
		t.Fatalf("expected 1 duplicate group, got %d", len(dupes))
	}
}
