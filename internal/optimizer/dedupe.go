package optimizer

import (
	"strings"

	"github.com/paulmach/orb/geojson"
)

// DeduplicateOptions controls deduplication behavior.
type DeduplicateOptions struct {
	ByID         bool
	ByCoordinate bool
	Precision    int
}

// Deduplicate removes duplicate features.
func Deduplicate(fc *geojson.FeatureCollection, opts DeduplicateOptions) *geojson.FeatureCollection {
	if opts.Precision <= 0 {
		opts.Precision = 6
	}
	seenID := make(map[string]struct{})
	seenCoord := make(map[string]struct{})
	result := make([]*geojson.Feature, 0, len(fc.Features))

	for _, f := range fc.Features {
		if f == nil {
			continue
		}
		if opts.ByID {
			id := featureID(f)
			if id != "" {
				if _, ok := seenID[id]; ok {
					continue
				}
				seenID[id] = struct{}{}
			}
		}
		if opts.ByCoordinate {
			key := coordKeyFromFeature(f)
			if key != "" {
				if _, ok := seenCoord[key]; ok {
					continue
				}
				seenCoord[key] = struct{}{}
			}
		}
		result = append(result, f)
	}
	return &geojson.FeatureCollection{Type: "FeatureCollection", Features: result}
}

// RemoveEmptyProperties strips null and empty string values.
func RemoveEmptyProperties(fc *geojson.FeatureCollection) *geojson.FeatureCollection {
	for _, f := range fc.Features {
		if f == nil || f.Properties == nil {
			continue
		}
		clean := make(map[string]any, len(f.Properties))
		for k, v := range f.Properties {
			if v == nil {
				continue
			}
			if s, ok := v.(string); ok && strings.TrimSpace(s) == "" {
				continue
			}
			clean[k] = v
		}
		f.Properties = clean
	}
	return fc
}

// SimplifyProperties drops and keeps fields per options.
func SimplifyProperties(fc *geojson.FeatureCollection, dropFields, keepFields []string) *geojson.FeatureCollection {
	drop := toSet(dropFields)
	keep := toSet(keepFields)
	useKeep := len(keep) > 0

	for _, f := range fc.Features {
		if f == nil || f.Properties == nil {
			continue
		}
		clean := make(map[string]any)
		for k, v := range f.Properties {
			if containsFold(drop, k) {
				continue
			}
			if useKeep && !containsFold(keep, k) {
				continue
			}
			clean[k] = v
		}
		f.Properties = clean
	}
	return fc
}

func toSet(fields []string) map[string]struct{} {
	s := make(map[string]struct{}, len(fields))
	for _, f := range fields {
		s[strings.ToLower(f)] = struct{}{}
	}
	return s
}

func containsFold(set map[string]struct{}, key string) bool {
	_, ok := set[strings.ToLower(key)]
	return ok
}

// CompressProperties keeps only essential fields when minimal mode is enabled.
func CompressProperties(fc *geojson.FeatureCollection) *geojson.FeatureCollection {
	essential := []string{"name", "title", "label", "type", "id", "description"}
	return SimplifyProperties(fc, nil, essential)
}
