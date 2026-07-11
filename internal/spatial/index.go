package spatial

import (
	"fmt"
	"math"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/quadtree"
)

// featurePointer wraps a feature for quadtree indexing.
type featurePointer struct {
	Index   int
	Feature *geojson.Feature
}

func (fp featurePointer) Point() orb.Point {
	if fp.Feature == nil {
		return orb.Point{math.NaN(), math.NaN()}
	}
	return fp.Feature.Point()
}

// Index provides spatial indexing over GeoJSON features.
type Index struct {
	tree     *quadtree.Quadtree
	features []featurePointer
	buf      []orb.Pointer
}

// NewIndex builds a quadtree index from features.
func NewIndex(fc *geojson.FeatureCollection) *Index {
	bounds := CollectionBound(fc)
	if bounds.IsEmpty() {
		bounds = orb.Bound{
			Min: orb.Point{-180, -90},
			Max: orb.Point{180, 90},
		}
	}
	tree := quadtree.New(bounds.Pad(0.01))

	idx := &Index{
		tree:     tree,
		features: make([]featurePointer, 0, len(fc.Features)),
		buf:      make([]orb.Pointer, 0, 64),
	}

	for i, f := range fc.Features {
		if f == nil || f.Geometry == nil {
			continue
		}
		fp := featurePointer{Index: i, Feature: f}
		idx.features = append(idx.features, fp)
		_ = tree.Add(fp)
	}
	return idx
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

// RadiusSearch returns features within radiusMeters of lat/lon.
func (idx *Index) RadiusSearch(lat, lon float64, radiusMeters float64) []*geojson.Feature {
	center := orb.Point{lon, lat}
	bbox := bboxFromRadius(center, radiusMeters)
	candidates := idx.tree.InBound(idx.buf[:0], bbox)

	result := make([]*geojson.Feature, 0)
	for _, c := range candidates {
		entry, ok := c.(featurePointer)
		if !ok {
			continue
		}
		pt := entry.Point()
		if HaversineMeters(lat, lon, pt[1], pt[0]) <= radiusMeters {
			result = append(result, entry.Feature)
		}
	}
	return result
}

// BBoxSearch returns features intersecting the bounding box [minLon, minLat, maxLon, maxLat].
func (idx *Index) BBoxSearch(minLon, minLat, maxLon, maxLat float64) []*geojson.Feature {
	bbox := orb.Bound{
		Min: orb.Point{minLon, minLat},
		Max: orb.Point{maxLon, maxLat},
	}
	candidates := idx.tree.InBound(idx.buf[:0], bbox)
	result := make([]*geojson.Feature, 0, len(candidates))
	for _, c := range candidates {
		entry, ok := c.(featurePointer)
		if !ok {
			continue
		}
		if bbox.Contains(entry.Point()) {
			result = append(result, entry.Feature)
		}
	}
	return result
}

// Nearest returns the nearest feature to lat/lon within maxMeters (0 = unlimited).
func (idx *Index) Nearest(lat, lon float64, maxMeters float64) *geojson.Feature {
	center := orb.Point{lon, lat}
	searchRadius := maxMeters
	if searchRadius <= 0 {
		searchRadius = 50000
	}
	bbox := bboxFromRadius(center, searchRadius)
	candidates := idx.tree.KNearest(idx.buf[:0], center, 1, searchRadius/111320.0)
	if len(candidates) == 0 {
		candidates = idx.tree.InBound(idx.buf[:0], bbox)
	}

	var nearest *geojson.Feature
	best := math.MaxFloat64
	for _, c := range candidates {
		entry, ok := c.(featurePointer)
		if !ok {
			continue
		}
		pt := entry.Point()
		d := HaversineMeters(lat, lon, pt[1], pt[0])
		if d < best {
			best = d
			nearest = entry.Feature
		}
	}
	if maxMeters > 0 && best > maxMeters {
		return nil
	}
	return nearest
}

// FindDuplicateCoordinates returns groups of feature indices sharing the same coordinate key.
func (idx *Index) FindDuplicateCoordinates(precision int) map[string][]int {
	if precision <= 0 {
		precision = 6
	}
	groups := make(map[string][]int)
	for _, entry := range idx.features {
		key := coordKey(entry.Point(), precision)
		groups[key] = append(groups[key], entry.Index)
	}
	dupes := make(map[string][]int)
	for k, v := range groups {
		if len(v) > 1 {
			dupes[k] = v
		}
	}
	return dupes
}

func coordKey(p orb.Point, precision int) string {
	factor := math.Pow(10, float64(precision))
	lon := math.Round(p[0]*factor) / factor
	lat := math.Round(p[1]*factor) / factor
	return fmt.Sprintf("%.6f,%.6f", lon, lat)
}

func bboxFromRadius(center orb.Point, radiusMeters float64) orb.Bound {
	latDelta := radiusMeters / 111320.0
	cosLat := math.Cos(center[1] * math.Pi / 180)
	if cosLat < 0.0001 {
		cosLat = 0.0001
	}
	lonDelta := radiusMeters / (111320.0 * cosLat)
	return orb.Bound{
		Min: orb.Point{center[0] - lonDelta, center[1] - latDelta},
		Max: orb.Point{center[0] + lonDelta, center[1] + latDelta},
	}
}
