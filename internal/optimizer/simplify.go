package optimizer

import (
	"fmt"
	"sort"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"

	"github.com/BaconFries/meshtastic-poi/internal/spatial"
)

// Options controls the optimization pipeline.
type Options struct {
	Minimal            bool
	RemoveEmpty        bool
	Dedupe             bool
	SortDistance       bool
	CompressProperties bool
	DropFields         []string
	KeepFields         []string
	SortLat            float64
	SortLon            float64
	CoordPrecision     int
}

// Pipeline runs the full optimization pipeline.
func Pipeline(fc *geojson.FeatureCollection, opts Options) *geojson.FeatureCollection {
	if fc == nil {
		return &geojson.FeatureCollection{Type: "FeatureCollection", Features: nil}
	}

	out := fc
	out = RemoveInvalidFeatures(out)
	out = NormalizeCoordinates(out, opts.CoordPrecision)

	if opts.RemoveEmpty || opts.Minimal {
		out = RemoveEmptyProperties(out)
	}

	if opts.CompressProperties || opts.Minimal {
		out = CompressProperties(out)
	}

	if len(opts.DropFields) > 0 || len(opts.KeepFields) > 0 {
		out = SimplifyProperties(out, opts.DropFields, opts.KeepFields)
	}

	if opts.Dedupe || opts.Minimal {
		out = Deduplicate(out, DeduplicateOptions{ByID: true, ByCoordinate: true, Precision: opts.CoordPrecision})
	}

	if opts.SortDistance && opts.SortLat != 0 && opts.SortLon != 0 {
		out = SortByDistance(out, opts.SortLat, opts.SortLon)
	} else {
		out = SortByID(out)
	}

	return out
}

// SortByDistance sorts features by distance from lat/lon.
func SortByDistance(fc *geojson.FeatureCollection, lat, lon float64) *geojson.FeatureCollection {
	ref := spatial.DistanceFrom{Lat: lat, Lon: lon}
	type item struct {
		f    *geojson.Feature
		dist float64
	}
	items := make([]item, 0, len(fc.Features))
	for _, f := range fc.Features {
		if f == nil || f.Geometry == nil {
			continue
		}
		c := f.Geometry.Bound().Center()
		items = append(items, item{f: f, dist: ref.DistanceTo(c)})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].dist < items[j].dist
	})
	features := make([]*geojson.Feature, len(items))
	for i, it := range items {
		features[i] = it.f
	}
	return &geojson.FeatureCollection{Type: "FeatureCollection", Features: features}
}

// SortByID sorts features by id or name property.
func SortByID(fc *geojson.FeatureCollection) *geojson.FeatureCollection {
	features := make([]*geojson.Feature, len(fc.Features))
	copy(features, fc.Features)
	sort.Slice(features, func(i, j int) bool {
		return sortKey(features[i]) < sortKey(features[j])
	})
	return &geojson.FeatureCollection{Type: "FeatureCollection", Features: features}
}

func sortKey(f *geojson.Feature) string {
	if f == nil {
		return ""
	}
	if id := featureID(f); id != "" {
		return id
	}
	if f.Properties != nil {
		for _, k := range []string{"name", "NAME", "title"} {
			if v, ok := f.Properties[k]; ok {
				return stringVal(v)
			}
		}
	}
	if f.Geometry != nil {
		c := f.Geometry.Bound().Center()
		return coordKeyFromPoint(c)
	}
	return ""
}

func stringVal(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func coordKeyFromPoint(p orb.Point) string {
	return fmt.Sprintf("%.6f,%.6f", p[0], p[1])
}
