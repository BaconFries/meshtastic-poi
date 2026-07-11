package exporters

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/paulmach/orb/geojson"

	"github.com/BaconFries/meshtastic-poi/internal/model"
)

// GeoJSON exports POIs as a GeoJSON FeatureCollection.
type GeoJSON struct{}

func (GeoJSON) Name() string { return "geojson" }

func (GeoJSON) Export(_ context.Context, pois []*model.POI, w io.Writer) error {
	fc := model.ToFeatureCollection(pois)
	data, err := json.MarshalIndent(fc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal geojson: %w", err)
	}
	_, err = w.Write(data)
	return err
}

// ReadGeoJSONFile loads POIs from a GeoJSON file.
func ReadGeoJSONFile(path string) ([]*model.POI, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var fc geojson.FeatureCollection
	if err := json.Unmarshal(data, &fc); err != nil {
		return nil, fmt.Errorf("parse geojson: %w", err)
	}
	return model.FromFeatureCollection(&fc, "geojson"), nil
}

// WriteGeoJSONFile writes POIs to a GeoJSON file or stdout.
func WriteGeoJSONFile(path string, pois []*model.POI) error {
	var w io.Writer = os.Stdout
	if path != "" && path != "-" {
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()
		w = f
	}
	return GeoJSON{}.Export(context.Background(), pois, w)
}

// Meshtastic exports POIs in Meshtastic node JSON format.
type Meshtastic struct{}

func (Meshtastic) Name() string { return "meshtastic" }

type meshtasticNode struct {
	Name        string  `json:"name"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Type        string  `json:"type,omitempty"`
	Description string  `json:"description,omitempty"`
}

type meshtasticExport struct {
	Nodes []meshtasticNode `json:"nodes"`
}

func (Meshtastic) Export(_ context.Context, pois []*model.POI, w io.Writer) error {
	nodes := make([]meshtasticNode, 0, len(pois))
	for _, p := range pois {
		if p == nil || !p.Valid() {
			continue
		}
		name := p.Name
		if name == "" {
			name = fmt.Sprintf("POI %.5f,%.5f", p.Location[1], p.Location[0])
		}
		nodes = append(nodes, meshtasticNode{
			Name:        name,
			Latitude:    p.Location[1],
			Longitude:   p.Location[0],
			Type:        p.Category,
			Description: p.Description,
		})
	}
	data, err := json.MarshalIndent(meshtasticExport{Nodes: nodes}, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal meshtastic: %w", err)
	}
	_, err = w.Write(data)
	return err
}

// CSV exports POIs as a CSV file with standard columns.
type CSV struct{}

func (CSV) Name() string { return "csv" }

func (CSV) Export(_ context.Context, pois []*model.POI, w io.Writer) error {
	cw := csv.NewWriter(w)
	if err := cw.Write([]string{"id", "name", "category", "description", "latitude", "longitude", "address", "source"}); err != nil {
		return err
	}
	for _, p := range pois {
		if p == nil {
			continue
		}
		if err := cw.Write([]string{
			p.ID,
			p.Name,
			p.Category,
			p.Description,
			fmt.Sprintf("%.7f", p.Location[1]),
			fmt.Sprintf("%.7f", p.Location[0]),
			p.Address,
			p.Source,
		}); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

// WriteFile writes POIs using a named exporter.
func WriteFile(path, format string, pois []*model.POI) error {
	reg := DefaultRegistry()
	exp, err := reg.Get(strings.ToLower(format))
	if err != nil {
		return err
	}
	var w io.Writer = os.Stdout
	if path != "" && path != "-" {
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()
		w = f
	}
	return exp.Export(context.Background(), pois, w)
}
