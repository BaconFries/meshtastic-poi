package optimizer

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/BaconFries/meshtastic-poi/internal/exporters"
	"github.com/BaconFries/meshtastic-poi/internal/model"
	"github.com/BaconFries/meshtastic-poi/internal/spatial"
	"github.com/paulmach/orb/geojson"
)

// Stats holds dataset statistics.
type Stats struct {
	FeatureCount           int     `json:"feature_count"`
	GeometryType           string  `json:"geometry_type"`
	AvgPropertyCount       float64 `json:"avg_property_count"`
	DuplicateCoordinates   int     `json:"duplicate_coordinates"`
	BBoxMinLon             float64 `json:"bbox_min_lon"`
	BBoxMinLat             float64 `json:"bbox_min_lat"`
	BBoxMaxLon             float64 `json:"bbox_max_lon"`
	BBoxMaxLat             float64 `json:"bbox_max_lat"`
	FileSizeBytes          int64   `json:"file_size_bytes"`
	EstimatedOptimizedSize int64   `json:"estimated_optimized_size"`
}

// ComputeStats calculates statistics for a FeatureCollection.
func ComputeStats(fc *geojson.FeatureCollection, fileSize int64) Stats {
	b := spatial.CollectionBound(fc)
	stats := Stats{
		FeatureCount:         len(fc.Features),
		GeometryType:         DominantGeometryType(fc),
		AvgPropertyCount:     AveragePropertyCount(fc),
		DuplicateCoordinates: CountDuplicateCoordinates(fc),
		BBoxMinLon:           b.Min[0],
		BBoxMinLat:           b.Min[1],
		BBoxMaxLon:           b.Max[0],
		BBoxMaxLat:           b.Max[1],
		FileSizeBytes:        fileSize,
	}
	if fileSize > 0 {
		optimized := Pipeline(fc, Options{Minimal: true, Dedupe: true, RemoveEmpty: true})
		data, err := json.Marshal(optimized)
		if err == nil {
			stats.EstimatedOptimizedSize = int64(len(data))
		} else {
			stats.EstimatedOptimizedSize = fileSize / 6
		}
	}
	return stats
}

// FormatStats returns human-readable stats output.
func FormatStats(s Stats) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Features\n\n%d\n\n", s.FeatureCount)
	fmt.Fprintf(&b, "Geometry\n\n%s\n\n", s.GeometryType)
	fmt.Fprintf(&b, "Properties\n\nAverage: %.0f\n\n", s.AvgPropertyCount)
	fmt.Fprintf(&b, "Duplicate Coordinates\n\n%d\n\n", s.DuplicateCoordinates)
	fmt.Fprintf(&b, "Bounding Box\n\n")
	fmt.Fprintf(&b, "  Min: %.6f, %.6f\n", s.BBoxMinLon, s.BBoxMinLat)
	fmt.Fprintf(&b, "  Max: %.6f, %.6f\n\n", s.BBoxMaxLon, s.BBoxMaxLat)
	if s.FileSizeBytes > 0 {
		fmt.Fprintf(&b, "File Size\n\nCurrent\n\n%s\n\n", formatBytes(s.FileSizeBytes))
		fmt.Fprintf(&b, "Estimated Optimized\n\n%s\n", formatBytes(s.EstimatedOptimizedSize))
	}
	return b.String()
}

func formatBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for n2 := n / unit; n2 >= unit; n2 /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}

// StatsFromFile loads a GeoJSON file and computes stats.
func StatsFromFile(path string) (Stats, error) {
	pois, err := exporters.ReadGeoJSONFile(path)
	if err != nil {
		return Stats{}, err
	}
	fc := model.ToFeatureCollection(pois)
	var fileSize int64
	if info, err := os.Stat(path); err == nil {
		fileSize = info.Size()
	}
	return ComputeStats(fc, fileSize), nil
}
