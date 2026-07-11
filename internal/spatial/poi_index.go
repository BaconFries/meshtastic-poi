package spatial

import (
	"fmt"
	"math"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/quadtree"

	"github.com/BaconFries/meshtastic-poi/internal/model"
)

// poiPointer wraps a POI for spatial indexing.
type poiPointer struct {
	Index int
	POI   *model.POI
}

func (pp poiPointer) Point() orb.Point {
	if pp.POI == nil {
		return orb.Point{math.NaN(), math.NaN()}
	}
	return pp.POI.Location
}

// POIIndex provides spatial indexing over canonical POIs.
type POIIndex struct {
	tree *quadtree.Quadtree
	pois []poiPointer
	buf  []orb.Pointer
}

// NewPOIIndex builds a spatial index from POIs.
func NewPOIIndex(pois []*model.POI) *POIIndex {
	bounds := POIBound(pois)
	if bounds.IsEmpty() {
		bounds = orb.Bound{
			Min: orb.Point{-180, -90},
			Max: orb.Point{180, 90},
		}
	}
	tree := quadtree.New(bounds.Pad(0.01))
	idx := &POIIndex{
		tree: tree,
		pois: make([]poiPointer, 0, len(pois)),
		buf:  make([]orb.Pointer, 0, 64),
	}
	for i, p := range pois {
		if p == nil || !p.Valid() {
			continue
		}
		pp := poiPointer{Index: i, POI: p}
		idx.pois = append(idx.pois, pp)
		_ = tree.Add(pp)
	}
	return idx
}

// POIBound returns the bounding box of all POI locations.
func POIBound(pois []*model.POI) orb.Bound {
	var b orb.Bound
	first := true
	for _, p := range pois {
		if p == nil || !p.Valid() {
			continue
		}
		pt := p.Location
		if first {
			b = orb.Bound{Min: pt, Max: pt}
			first = false
			continue
		}
		if pt[0] < b.Min[0] {
			b.Min[0] = pt[0]
		}
		if pt[1] < b.Min[1] {
			b.Min[1] = pt[1]
		}
		if pt[0] > b.Max[0] {
			b.Max[0] = pt[0]
		}
		if pt[1] > b.Max[1] {
			b.Max[1] = pt[1]
		}
	}
	return b
}

// RadiusSearch returns POIs within radiusMeters of lat/lon.
func (idx *POIIndex) RadiusSearch(lat, lon float64, radiusMeters float64) []*model.POI {
	center := orb.Point{lon, lat}
	bbox := bboxFromRadius(center, radiusMeters)
	candidates := idx.tree.InBound(idx.buf[:0], bbox)
	result := make([]*model.POI, 0)
	for _, c := range candidates {
		entry, ok := c.(poiPointer)
		if !ok {
			continue
		}
		pt := entry.Point()
		if HaversineMeters(lat, lon, pt[1], pt[0]) <= radiusMeters {
			result = append(result, entry.POI)
		}
	}
	return result
}

// BBoxSearch returns POIs inside the bounding box.
func (idx *POIIndex) BBoxSearch(minLon, minLat, maxLon, maxLat float64) []*model.POI {
	bbox := orb.Bound{
		Min: orb.Point{minLon, minLat},
		Max: orb.Point{maxLon, maxLat},
	}
	candidates := idx.tree.InBound(idx.buf[:0], bbox)
	result := make([]*model.POI, 0, len(candidates))
	for _, c := range candidates {
		entry, ok := c.(poiPointer)
		if !ok {
			continue
		}
		if bbox.Contains(entry.Point()) {
			result = append(result, entry.POI)
		}
	}
	return result
}

// Nearest returns the nearest POI to lat/lon within maxMeters (0 = unlimited).
func (idx *POIIndex) Nearest(lat, lon float64, maxMeters float64) *model.POI {
	center := orb.Point{lon, lat}
	searchRadius := maxMeters
	if searchRadius <= 0 {
		searchRadius = 50000
	}
	candidates := idx.tree.KNearest(idx.buf[:0], center, 1, searchRadius/111320.0)
	if len(candidates) == 0 {
		bbox := bboxFromRadius(center, searchRadius)
		candidates = idx.tree.InBound(idx.buf[:0], bbox)
	}
	var nearest *model.POI
	best := math.MaxFloat64
	for _, c := range candidates {
		entry, ok := c.(poiPointer)
		if !ok {
			continue
		}
		pt := entry.Point()
		d := HaversineMeters(lat, lon, pt[1], pt[0])
		if d < best {
			best = d
			nearest = entry.POI
		}
	}
	if maxMeters > 0 && best > maxMeters {
		return nil
	}
	return nearest
}

// FindDuplicateCoordinates returns groups of POI indices sharing coordinates.
func (idx *POIIndex) FindDuplicateCoordinates(precision int) map[string][]int {
	if precision <= 0 {
		precision = 6
	}
	groups := make(map[string][]int)
	for _, entry := range idx.pois {
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

// FilterPOIRadius returns POIs within radius of a point.
func FilterPOIRadius(pois []*model.POI, lat, lon, radiusMeters float64) []*model.POI {
	idx := NewPOIIndex(pois)
	return idx.RadiusSearch(lat, lon, radiusMeters)
}

// FilterPOIBBox returns POIs inside a bounding box.
func FilterPOIBBox(pois []*model.POI, minLon, minLat, maxLon, maxLat float64) []*model.POI {
	idx := NewPOIIndex(pois)
	return idx.BBoxSearch(minLon, minLat, maxLon, maxLat)
}

// FilterPOIAttributes returns POIs whose tags match attribute filters.
func FilterPOIAttributes(pois []*model.POI, filters map[string]string) []*model.POI {
	if len(filters) == 0 {
		return pois
	}
	out := make([]*model.POI, 0)
	for _, p := range pois {
		if p == nil {
			continue
		}
		match := true
		for k, want := range filters {
			got, ok := p.Tags[k]
			if !ok {
				got = fmt.Sprint(p.Metadata[k])
			}
			if !containsFold(got, want) {
				match = false
				break
			}
		}
		if match {
			out = append(out, p)
		}
	}
	return out
}

func containsFold(haystack, needle string) bool {
	return len(needle) == 0 || (len(haystack) >= len(needle) &&
		(haystack == needle || len(needle) > 0 && stringContainsFold(haystack, needle)))
}

func stringContainsFold(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) &&
		(s == sub || indexFold(s, sub) >= 0))
}

func indexFold(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if equalFold(s[i:i+len(sub)], sub) {
			return i
		}
	}
	return -1
}

func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}
