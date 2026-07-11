package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
)

var nameKeys = []string{"name", "title", "label", "NAME", "Title"}
var categoryKeys = []string{"type", "category", "TYPE", "Category", "class"}
var descKeys = []string{"description", "desc", "notes", "comment"}
var addressKeys = []string{"address", "addr", "street", "full_address"}
var idKeys = []string{"id", "ID", "OBJECTID", "objectid", "fid", "FID", "osm_id"}

// FromFeature converts a GeoJSON feature into a POI.
func FromFeature(f *geojson.Feature, source string) *POI {
	if f == nil {
		return nil
	}
	poi := &POI{
		Tags:     make(map[string]string),
		Metadata: make(map[string]any),
		Source:   source,
	}
	poi.ID = featureID(f)
	poi.Name = propertyString(f.Properties, nameKeys...)
	poi.Category = propertyString(f.Properties, categoryKeys...)
	poi.Description = propertyString(f.Properties, descKeys...)
	poi.Address = propertyString(f.Properties, addressKeys...)
	if f.Geometry != nil {
		poi.Location = locationFromGeometry(f.Geometry)
		if _, isPoint := f.Geometry.(orb.Point); !isPoint {
			poi.Metadata["geometry"] = f.Geometry
			poi.Metadata["geometry_type"] = f.Geometry.GeoJSONType()
		}
	}
	if f.Properties != nil {
		for k, v := range f.Properties {
			if isReservedProperty(k) {
				continue
			}
			if s, ok := v.(string); ok {
				poi.Tags[k] = s
				continue
			}
			poi.Metadata[k] = v
		}
	}
	if updated, ok := parseUpdated(poi.Tags, poi.Metadata); ok {
		poi.Updated = updated
	}
	return poi
}

// FromFeatureCollection converts all features in a collection.
func FromFeatureCollection(fc *geojson.FeatureCollection, source string) []*POI {
	if fc == nil {
		return nil
	}
	out := make([]*POI, 0, len(fc.Features))
	for _, f := range fc.Features {
		if p := FromFeature(f, source); p != nil {
			out = append(out, p)
		}
	}
	return out
}

// ToFeature converts a POI into a GeoJSON feature.
func ToFeature(p *POI) *geojson.Feature {
	if p == nil {
		return nil
	}
	props := geojson.Properties{}
	if p.Name != "" {
		props["name"] = p.Name
	}
	if p.Category != "" {
		props["type"] = p.Category
	}
	if p.Description != "" {
		props["description"] = p.Description
	}
	if p.Address != "" {
		props["address"] = p.Address
	}
	if p.Source != "" {
		props["source"] = p.Source
	}
	for k, v := range p.Tags {
		props[k] = v
	}
	for k, v := range p.Metadata {
		if k == "geometry" || k == "geometry_type" {
			continue
		}
		props[k] = v
	}
	var geom orb.Geometry = p.Location
	if g, ok := p.Metadata["geometry"].(orb.Geometry); ok && g != nil {
		geom = g
	}
	f := geojson.NewFeature(geom)
	f.Properties = props
	if p.ID != "" {
		f.ID = p.ID
	}
	return f
}

// ToFeatureCollection converts POIs to a GeoJSON feature collection.
func ToFeatureCollection(pois []*POI) *geojson.FeatureCollection {
	features := make([]*geojson.Feature, 0, len(pois))
	for _, p := range pois {
		if f := ToFeature(p); f != nil {
			features = append(features, f)
		}
	}
	return &geojson.FeatureCollection{Type: "FeatureCollection", Features: features}
}

func featureID(f *geojson.Feature) string {
	if f.ID != nil {
		switch id := f.ID.(type) {
		case string:
			return id
		case float64:
			return fmt.Sprintf("%.0f", id)
		default:
			return fmt.Sprintf("%v", id)
		}
	}
	return propertyString(f.Properties, idKeys...)
}

func propertyString(props geojson.Properties, keys ...string) string {
	if props == nil {
		return ""
	}
	for _, key := range keys {
		for k, v := range props {
			if strings.EqualFold(k, key) {
				if s, ok := v.(string); ok {
					return s
				}
				return fmt.Sprintf("%v", v)
			}
		}
	}
	return ""
}

func locationFromGeometry(g orb.Geometry) orb.Point {
	if p, ok := g.(orb.Point); ok {
		return p
	}
	b := g.Bound()
	if !b.IsEmpty() {
		return b.Center()
	}
	return orb.Point{}
}

func isReservedProperty(key string) bool {
	switch strings.ToLower(key) {
	case "name", "title", "label", "type", "category", "description", "desc", "address", "id", "objectid", "fid":
		return true
	}
	return false
}

func parseUpdated(tags map[string]string, meta map[string]any) (time.Time, bool) {
	for _, key := range []string{"updated", "last_updated", "modified", "timestamp"} {
		if v, ok := tags[key]; ok {
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				return t, true
			}
		}
		if v, ok := meta[key]; ok {
			switch t := v.(type) {
			case string:
				if parsed, err := time.Parse(time.RFC3339, t); err == nil {
					return parsed, true
				}
			case time.Time:
				return t, true
			}
		}
	}
	return time.Time{}, false
}
