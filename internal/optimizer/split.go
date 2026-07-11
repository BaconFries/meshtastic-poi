package optimizer

import (
	"fmt"
	"math"
	"strings"

	"github.com/paulmach/orb/geojson"
)

// SplitOptions controls dataset splitting.
type SplitOptions struct {
	ByField     string
	MaxFeatures int
	TileSizeDeg float64
}

// Split divides a FeatureCollection into multiple collections.
func Split(fc *geojson.FeatureCollection, opts SplitOptions) (map[string]*geojson.FeatureCollection, error) {
	if opts.ByField != "" {
		return splitByField(fc, opts.ByField), nil
	}
	if opts.MaxFeatures > 0 {
		return splitByCount(fc, opts.MaxFeatures), nil
	}
	if opts.TileSizeDeg > 0 {
		return splitByTile(fc, opts.TileSizeDeg), nil
	}
	return map[string]*geojson.FeatureCollection{"all": fc}, nil
}

func splitByField(fc *geojson.FeatureCollection, field string) map[string]*geojson.FeatureCollection {
	groups := make(map[string]*geojson.FeatureCollection)
	for _, f := range fc.Features {
		key := "unknown"
		if f != nil && f.Properties != nil {
			for k, v := range f.Properties {
				if equalFold(k, field) {
					key = sanitizeKey(fmt.Sprintf("%v", v))
					break
				}
			}
		}
		if groups[key] == nil {
			groups[key] = &geojson.FeatureCollection{Type: "FeatureCollection", Features: make([]*geojson.Feature, 0)}
		}
		groups[key].Features = append(groups[key].Features, f)
	}
	return groups
}

func splitByCount(fc *geojson.FeatureCollection, max int) map[string]*geojson.FeatureCollection {
	groups := make(map[string]*geojson.FeatureCollection)
	chunk := 0
	for i, f := range fc.Features {
		if i%max == 0 {
			chunk = i / max
			key := fmt.Sprintf("part_%04d", chunk)
			groups[key] = &geojson.FeatureCollection{Type: "FeatureCollection", Features: make([]*geojson.Feature, 0)}
		}
		key := fmt.Sprintf("part_%04d", chunk)
		groups[key].Features = append(groups[key].Features, f)
	}
	return groups
}

func splitByTile(fc *geojson.FeatureCollection, size float64) map[string]*geojson.FeatureCollection {
	groups := make(map[string]*geojson.FeatureCollection)
	for _, f := range fc.Features {
		if f == nil || f.Geometry == nil {
			continue
		}
		c := f.Geometry.Bound().Center()
		tileX := int(math.Floor(c[0] / size))
		tileY := int(math.Floor(c[1] / size))
		key := fmt.Sprintf("tile_%d_%d", tileX, tileY)
		if groups[key] == nil {
			groups[key] = &geojson.FeatureCollection{Type: "FeatureCollection", Features: make([]*geojson.Feature, 0)}
		}
		groups[key].Features = append(groups[key].Features, f)
	}
	return groups
}

func sanitizeKey(s string) string {
	replacer := strings.NewReplacer("/", "_", "\\", "_", " ", "_", ":", "_")
	return replacer.Replace(s)
}

func equalFold(a, b string) bool {
	return strings.EqualFold(a, b)
}
