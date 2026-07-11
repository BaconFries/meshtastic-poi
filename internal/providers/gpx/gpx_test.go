package gpx_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BaconFries/meshtastic-poi/internal/providers/gpx"
)

func TestParseGPX(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "sample.gpx"))
	if err != nil {
		t.Fatal(err)
	}
	fc, err := gpx.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(fc.Features) != 3 {
		t.Fatalf("expected 3 features, got %d", len(fc.Features))
	}
}
