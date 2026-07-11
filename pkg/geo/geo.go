package geo

import (
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
)

// FeatureCount returns the number of features in a collection.
func FeatureCount(fc *geojson.FeatureCollection) int {
	if fc == nil {
		return 0
	}
	return len(fc.Features)
}

// CollectionBound returns the bounding box of all features.
func CollectionBound(fc *geojson.FeatureCollection) orb.Bound {
	var b orb.Bound
	first := true
	for _, f := range fc.Features {
		if f == nil || f.Geometry == nil {
			continue
		}
		fb := f.Geometry.Bound()
		if fb.IsEmpty() {
			continue
		}
		if first {
			b = fb
			first = false
			continue
		}
		b = b.Union(fb)
	}
	return b
}
