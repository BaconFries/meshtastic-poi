package optimizer

import (
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
)

// NormalizeCoordinates rounds coordinates to the given decimal precision (default 7).
func NormalizeCoordinates(fc *geojson.FeatureCollection, precision int) *geojson.FeatureCollection {
	if precision <= 0 {
		precision = 7
	}
	for _, f := range fc.Features {
		if f == nil || f.Geometry == nil {
			continue
		}
		f.Geometry = orb.Round(f.Geometry, precision)
	}
	return fc
}
