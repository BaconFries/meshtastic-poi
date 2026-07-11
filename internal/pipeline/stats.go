package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/BaconFries/meshtastic-poi/internal/exporters"
	"github.com/BaconFries/meshtastic-poi/internal/model"
	"github.com/BaconFries/meshtastic-poi/internal/spatial"
)

// Stats holds dataset statistics for canonical POIs.
type Stats struct {
	POICount               int     `json:"poi_count"`
	AvgTagCount            float64 `json:"avg_tag_count"`
	DuplicateCoordinates   int     `json:"duplicate_coordinates"`
	BBoxMinLon             float64 `json:"bbox_min_lon"`
	BBoxMinLat             float64 `json:"bbox_min_lat"`
	BBoxMaxLon             float64 `json:"bbox_max_lon"`
	BBoxMaxLat             float64 `json:"bbox_max_lat"`
	FileSizeBytes          int64   `json:"file_size_bytes"`
	EstimatedOptimizedSize int64   `json:"estimated_optimized_size"`
}

// ComputeStats calculates statistics for POIs.
func ComputeStats(pois []*model.POI, fileSize int64) Stats {
	b := spatial.POIBound(pois)
	idx := spatial.NewPOIIndex(pois)
	stats := Stats{
		POICount:             len(pois),
		AvgTagCount:          averageTagCount(pois),
		DuplicateCoordinates: len(idx.FindDuplicateCoordinates(6)),
		BBoxMinLon:           b.Min[0],
		BBoxMinLat:           b.Min[1],
		BBoxMaxLon:           b.Max[0],
		BBoxMaxLat:           b.Max[1],
		FileSizeBytes:        fileSize,
	}
	if fileSize > 0 {
		optimized, _ := Run(context.Background(), pois, Default(Options{Minimal: true, Dedupe: true, RemoveEmpty: true}))
		fc := model.ToFeatureCollection(optimized)
		data, err := json.Marshal(fc)
		if err == nil {
			stats.EstimatedOptimizedSize = int64(len(data))
		}
	}
	return stats
}

// StatsFromFile loads GeoJSON and computes POI stats.
func StatsFromFile(path string) (Stats, error) {
	pois, err := exporters.ReadGeoJSONFile(path)
	if err != nil {
		return Stats{}, err
	}
	var fileSize int64
	if info, err := os.Stat(path); err == nil {
		fileSize = info.Size()
	}
	return ComputeStats(pois, fileSize), nil
}

// FormatStats returns human-readable stats output.
func FormatStats(s Stats) string {
	var b strings.Builder
	fmt.Fprintf(&b, "POIs\n\n%d\n\n", s.POICount)
	fmt.Fprintf(&b, "Tags\n\nAverage: %.0f\n\n", s.AvgTagCount)
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

func averageTagCount(pois []*model.POI) float64 {
	if len(pois) == 0 {
		return 0
	}
	total := 0
	for _, p := range pois {
		if p != nil && p.Tags != nil {
			total += len(p.Tags)
		}
	}
	return float64(total) / float64(len(pois))
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
