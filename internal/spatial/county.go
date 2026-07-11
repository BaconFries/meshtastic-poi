package spatial

import (
	"strings"

	"github.com/paulmach/orb/geojson"
)

// SplitByField groups features by a property field value.
func SplitByField(fc *geojson.FeatureCollection, field string) map[string]*geojson.FeatureCollection {
	groups := make(map[string]*geojson.FeatureCollection)
	for _, f := range fc.Features {
		if f == nil {
			continue
		}
		key := "_unknown_"
		if f.Properties != nil {
			if v, ok := findProperty(f.Properties, field); ok && v != "" {
				key = sanitizeFilename(v)
			}
		}
		if groups[key] == nil {
			groups[key] = &geojson.FeatureCollection{Type: "FeatureCollection", Features: make([]*geojson.Feature, 0)}
		}
		groups[key].Features = append(groups[key].Features, f)
	}
	return groups
}

func sanitizeFilename(s string) string {
	replacer := strings.NewReplacer(
		"/", "_", "\\", "_", ":", "_", "*", "_",
		"?", "_", "\"", "_", "<", "_", ">", "_", "|", "_",
		" ", "_",
	)
	return replacer.Replace(s)
}
