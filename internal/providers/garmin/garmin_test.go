package garmin_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BaconFries/meshtastic-poi/internal/providers/garmin"
)

func TestParseGarminCSV(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "sample_garmin.csv"))
	if err != nil {
		t.Fatal(err)
	}
	fc, err := garmin.ParseCSV(data, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(fc.Features) != 3 {
		t.Fatalf("expected 3 features, got %d", len(fc.Features))
	}
	if fc.Features[0].Properties["name"] != "Orlando POI" {
		t.Fatalf("unexpected name: %v", fc.Features[0].Properties["name"])
	}
}
