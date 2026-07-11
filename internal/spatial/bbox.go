package spatial

import (
	"fmt"
	"strings"

	"github.com/paulmach/orb/geojson"
)

// FilterAttributes returns features matching all provided attribute filters (case-insensitive substring).
func FilterAttributes(fc *geojson.FeatureCollection, filters map[string]string) *geojson.FeatureCollection {
	if len(filters) == 0 {
		return fc
	}
	result := make([]*geojson.Feature, 0, len(fc.Features))
	for _, f := range fc.Features {
		if f == nil || f.Properties == nil {
			continue
		}
		if matchesAll(f.Properties, filters) {
			result = append(result, f)
		}
	}
	return &geojson.FeatureCollection{Type: "FeatureCollection", Features: result}
}

func matchesAll(props map[string]any, filters map[string]string) bool {
	for key, want := range filters {
		if want == "" {
			continue
		}
		val, ok := findProperty(props, key)
		if !ok {
			return false
		}
		if !strings.Contains(strings.ToLower(val), strings.ToLower(want)) {
			return false
		}
	}
	return true
}

func findProperty(props map[string]any, key string) (string, bool) {
	for k, v := range props {
		if strings.EqualFold(k, key) {
			switch t := v.(type) {
			case string:
				return t, true
			case float64:
				return strings.TrimRight(strings.TrimRight(fmtFloat(t), "0"), "."), true
			default:
				return fmt.Sprintf("%v", v), true
			}
		}
	}
	return "", false
}

func fmtFloat(f float64) string {
	return strings.TrimSpace(strings.TrimRight(strings.TrimRight(
		strings.Replace(fmt.Sprintf("%.6f", f), "-0.000000", "0.000000", 1),
		"0"), "."))
}
