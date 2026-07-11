package spatial

import (
	"fmt"
	"math"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
)

const earthRadiusMeters = 6371000

// HaversineMeters returns the great-circle distance in meters between two WGS84 points.
func HaversineMeters(lat1, lon1, lat2, lon2 float64) float64 {
	rad := math.Pi / 180
	φ1 := lat1 * rad
	φ2 := lat2 * rad
	dφ := (lat2 - lat1) * rad
	dλ := (lon2 - lon1) * rad

	a := math.Sin(dφ/2)*math.Sin(dφ/2) +
		math.Cos(φ1)*math.Cos(φ2)*math.Sin(dλ/2)*math.Sin(dλ/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusMeters * c
}

// FilterRadius returns features whose centroid is within radiusMeters of lat/lon.
func FilterRadius(fc *geojson.FeatureCollection, lat, lon, radiusMeters float64) *geojson.FeatureCollection {
	idx := NewIndex(fc)
	features := idx.RadiusSearch(lat, lon, radiusMeters)
	return &geojson.FeatureCollection{Type: "FeatureCollection", Features: features}
}

// ParseBBox parses "minLon,minLat,maxLon,maxLat".
func ParseBBox(s string) ([4]float64, error) {
	var bbox [4]float64
	n, err := fmt.Sscanf(s, "%f,%f,%f,%f", &bbox[0], &bbox[1], &bbox[2], &bbox[3])
	if err != nil || n != 4 {
		return bbox, fmt.Errorf("invalid bbox %q: expected minLon,minLat,maxLon,maxLat", s)
	}
	return bbox, nil
}

// FilterBBox returns features within the bounding box.
func FilterBBox(fc *geojson.FeatureCollection, minLon, minLat, maxLon, maxLat float64) *geojson.FeatureCollection {
	idx := NewIndex(fc)
	features := idx.BBoxSearch(minLon, minLat, maxLon, maxLat)
	return &geojson.FeatureCollection{Type: "FeatureCollection", Features: features}
}

// DistanceFrom sorts reference point for distance-based ordering.
type DistanceFrom struct {
	Lat float64
	Lon float64
}

// DistanceTo returns meters from reference to point.
func (d DistanceFrom) DistanceTo(p orb.Point) float64 {
	return HaversineMeters(d.Lat, d.Lon, p[1], p[0])
}
