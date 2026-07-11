package kml_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BaconFries/meshtastic-poi/internal/providers/kml"
)

func TestParseKML(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "sample.kml"))
	if err != nil {
		t.Fatal(err)
	}
	fc, err := kml.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(fc.Features) != 3 {
		t.Fatalf("expected 3 features, got %d", len(fc.Features))
	}
}
