package osm_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BaconFries/meshtastic-poi/internal/providers"
	"github.com/BaconFries/meshtastic-poi/internal/providers/osm"
)

func TestParseOverpassJSON(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "sample_overpass.json"))
	if err != nil {
		t.Fatal(err)
	}
	fc, err := osm.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(fc.Features) != 2 {
		t.Fatalf("expected 2 features, got %d", len(fc.Features))
	}
}

func TestOverpassDownload(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "sample_overpass.json"))
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "expected POST", http.StatusMethodNotAllowed)
			return
		}
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	cfg := providers.SourceConfig{
		Name: "test-osm",
		Type: "osm",
		URL:  srv.URL,
		Params: map[string]string{
			"query": `[out:json];node["amenity"="park"];out;`,
		},
	}
	p, err := osm.New(cfg, providers.Dependencies{})
	if err != nil {
		t.Fatal(err)
	}
	fc, err := p.Download(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(fc.Features) != 2 {
		t.Fatalf("expected 2 features, got %d", len(fc.Features))
	}
}

func TestBuildBBoxQuery(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		query := r.FormValue("data")
		if !strings.Contains(query, `["amenity"="park"]`) {
			t.Errorf("expected amenity filter in query, got: %s", query)
		}
		if !strings.Contains(query, "(28.000000,-82.000000,29.000000,-81.000000)") {
			t.Errorf("expected bbox in query, got: %s", query)
		}
		_, _ = w.Write([]byte(`{"elements":[]}`))
	}))
	defer srv.Close()

	cfg := providers.SourceConfig{
		URL:  srv.URL,
		BBox: []float64{-82.0, 28.0, -81.0, 29.0},
		Params: map[string]string{
			"tags": "amenity=park",
		},
	}
	p, err := osm.New(cfg, providers.Dependencies{})
	if err != nil {
		t.Fatal(err)
	}
	fc, err := p.Download(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(fc.Features) != 0 {
		t.Fatalf("expected empty result, got %d", len(fc.Features))
	}
}
