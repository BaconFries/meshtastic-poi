package arcgis_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/BaconFries/meshtastic-poi/internal/providers"
	"github.com/BaconFries/meshtastic-poi/internal/providers/arcgis"
)

func TestArcGISPagination(t *testing.T) {
	page := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/layer":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"name": "Test Layer",
				"maxRecordCount": 2,
				"geometryType": "esriGeometryPoint",
				"fields": [{"name": "OBJECTID", "type": "esriFieldTypeOID"}]
			}`))
		case r.URL.Path == "/layer/query":
			page++
			w.Header().Set("Content-Type", "application/json")
			if page == 1 {
				_, _ = w.Write([]byte(`{
					"type": "FeatureCollection",
					"features": [
						{"type":"Feature","geometry":{"type":"Point","coordinates":[-81,28]},"properties":{"OBJECTID":1}},
						{"type":"Feature","geometry":{"type":"Point","coordinates":[-82,29]},"properties":{"OBJECTID":2}}
					],
					"exceededTransferLimit": true
				}`))
				return
			}
			_, _ = w.Write([]byte(`{
				"type": "FeatureCollection",
				"features": [
					{"type":"Feature","geometry":{"type":"Point","coordinates":[-83,30]},"properties":{"OBJECTID":3}}
				],
				"exceededTransferLimit": false
			}`))
		}
	}))
	defer srv.Close()

	cfg := providers.SourceConfig{
		Name: "test",
		Type: "arcgis",
		URL:  srv.URL + "/layer",
	}
	p, err := arcgis.New(cfg, providers.Dependencies{})
	if err != nil {
		t.Fatal(err)
	}

	pois, err := p.Fetch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(pois) != 3 {
		t.Fatalf("expected 3 POIs across pages, got %d", len(pois))
	}

	meta, err := p.Metadata(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if meta.MaxRecordCount != 2 {
		t.Fatalf("expected maxRecordCount 2, got %d", meta.MaxRecordCount)
	}
}

func TestEsriGeometryConversion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/layer" && r.URL.RawQuery == "f=json" {
			_, _ = w.Write([]byte(`{"name":"t","maxRecordCount":1000,"geometryType":"esriGeometryPoint","fields":[]}`))
			return
		}
		if r.URL.Path == "/layer/query" {
			_, _ = w.Write([]byte(`{
				"features": [{
					"attributes": {"name": "Test"},
					"geometry": {"x": -81.5, "y": 28.5}
				}],
				"exceededTransferLimit": false
			}`))
		}
	}))
	defer srv.Close()

	cfg := providers.SourceConfig{
		Name: "test",
		Type: "arcgis",
		URL:  srv.URL + "/layer",
	}
	p, err := arcgis.New(cfg, providers.Dependencies{})
	if err != nil {
		t.Fatal(err)
	}
	pois, err := p.Fetch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(pois) != 1 {
		t.Fatalf("expected 1 POI, got %d", len(pois))
	}
}
