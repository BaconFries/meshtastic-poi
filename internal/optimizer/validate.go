package optimizer

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"

	"github.com/BaconFries/meshtastic-poi/internal/spatial"
)

// IssueSeverity classifies validation problems.
type IssueSeverity string

const (
	SeverityError   IssueSeverity = "error"
	SeverityWarning IssueSeverity = "warning"
)

// Issue describes a single validation problem.
type Issue struct {
	Severity IssueSeverity `json:"severity"`
	Code     string        `json:"code"`
	Message  string        `json:"message"`
	Index    int           `json:"index,omitempty"`
	Field    string        `json:"field,omitempty"`
}

// Report aggregates validation results.
type Report struct {
	Valid           bool    `json:"valid"`
	FeatureCount    int     `json:"feature_count"`
	ValidFeatures   int     `json:"valid_features"`
	InvalidFeatures int     `json:"invalid_features"`
	Issues          []Issue `json:"issues"`
}

// Validate checks a FeatureCollection and returns a report.
func Validate(fc *geojson.FeatureCollection) Report {
	report := Report{
		Valid:        true,
		FeatureCount: len(fc.Features),
		Issues:       make([]Issue, 0),
	}

	seenIDs := make(map[string]int)
	coordKeys := make(map[string]int)

	for i, f := range fc.Features {
		if f == nil {
			report.add(SeverityError, "null_feature", "feature is nil", i, "")
			continue
		}

		if f.Geometry == nil {
			report.add(SeverityError, "null_geometry", "geometry is null", i, "")
			continue
		}

		if err := validateGeometry(f.Geometry, i, &report); err != nil {
			continue
		}

		validateProperties(f.Properties, i, &report)

		if id := featureID(f); id != "" {
			if prev, ok := seenIDs[id]; ok {
				report.add(SeverityWarning, "duplicate_id",
					fmt.Sprintf("duplicate id %q (also at index %d)", id, prev), i, "id")
			}
			seenIDs[id] = i
		}

		key := coordKeyFromFeature(f)
		if key != "" {
			if prev, ok := coordKeys[key]; ok {
				report.add(SeverityWarning, "duplicate_coordinates",
					fmt.Sprintf("duplicate coordinates (also at index %d)", prev), i, "")
			}
			coordKeys[key] = i
		}

		report.ValidFeatures++
	}

	report.InvalidFeatures = report.FeatureCount - report.ValidFeatures
	if report.InvalidFeatures > 0 {
		report.Valid = false
	}
	return report
}

func (r *Report) add(sev IssueSeverity, code, msg string, idx int, field string) {
	r.Issues = append(r.Issues, Issue{
		Severity: sev,
		Code:     code,
		Message:  msg,
		Index:    idx,
		Field:    field,
	})
}

func validateGeometry(g orb.Geometry, idx int, report *Report) error {
	geomType := g.GeoJSONType()
	if geomType == "" {
		report.add(SeverityError, "missing_geometry_type", "geometry type is empty", idx, "")
		return fmt.Errorf("missing type")
	}

	switch geomType {
	case "Point", "MultiPoint", "LineString", "MultiLineString", "Polygon", "MultiPolygon":
	default:
		report.add(SeverityWarning, "unsupported_geometry",
			fmt.Sprintf("unsupported geometry type %q", geomType), idx, "")
	}

	if hasNaNCoords(g) {
		report.add(SeverityError, "nan_coordinates", "geometry contains NaN coordinates", idx, "")
		return fmt.Errorf("nan coords")
	}

	if isEmptyGeometry(g) {
		report.add(SeverityError, "empty_coordinates", "geometry has empty coordinates", idx, "")
		return fmt.Errorf("empty coords")
	}
	return nil
}

func validateProperties(props geojson.Properties, idx int, report *Report) {
	if props == nil {
		return
	}
	for k, v := range props {
		if v == nil {
			report.add(SeverityWarning, "null_property",
				fmt.Sprintf("property %q is null", k), idx, k)
		}
	}
}

func featureID(f *geojson.Feature) string {
	if f.ID != nil {
		switch id := f.ID.(type) {
		case string:
			return id
		case float64:
			return fmt.Sprintf("%.0f", id)
		default:
			return fmt.Sprintf("%v", id)
		}
	}
	if f.Properties != nil {
		for _, key := range []string{"id", "ID", "OBJECTID", "objectid", "fid", "FID"} {
			if v, ok := f.Properties[key]; ok {
				return fmt.Sprintf("%v", v)
			}
		}
	}
	return ""
}

func coordKeyFromFeature(f *geojson.Feature) string {
	if f == nil || f.Geometry == nil {
		return ""
	}
	b := f.Geometry.Bound()
	if b.IsEmpty() {
		return ""
	}
	c := b.Center()
	return fmt.Sprintf("%.6f,%.6f", c[0], c[1])
}

func hasNaNCoords(g orb.Geometry) bool {
	return walkCoords(g, func(p orb.Point) bool {
		return math.IsNaN(p[0]) || math.IsNaN(p[1]) || math.IsInf(p[0], 0) || math.IsInf(p[1], 0)
	})
}

func isEmptyGeometry(g orb.Geometry) bool {
	if g == nil {
		return true
	}
	return g.Bound().IsEmpty()
}

func walkCoords(g orb.Geometry, fn func(orb.Point) bool) bool {
	if g == nil {
		return false
	}
	switch geom := g.(type) {
	case orb.Point:
		return fn(geom)
	case orb.MultiPoint:
		for _, p := range geom {
			if fn(p) {
				return true
			}
		}
	case orb.LineString:
		for _, p := range geom {
			if fn(p) {
				return true
			}
		}
	case orb.MultiLineString:
		for _, ls := range geom {
			for _, p := range ls {
				if fn(p) {
					return true
				}
			}
		}
	case orb.Polygon:
		for _, ring := range geom {
			for _, p := range ring {
				if fn(p) {
					return true
				}
			}
		}
	case orb.MultiPolygon:
		for _, poly := range geom {
			for _, ring := range poly {
				for _, p := range ring {
					if fn(p) {
						return true
					}
				}
			}
		}
	}
	return false
}

// ValidateJSON parses and validates raw GeoJSON bytes.
func ValidateJSON(data []byte) (Report, error) {
	var fc geojson.FeatureCollection
	if err := json.Unmarshal(data, &fc); err != nil {
		return Report{Valid: false, Issues: []Issue{{
			Severity: SeverityError,
			Code:     "invalid_geojson",
			Message:  err.Error(),
		}}}, err
	}
	return Validate(&fc), nil
}

// RemoveInvalidFeatures returns only features that pass basic geometry validation.
func RemoveInvalidFeatures(fc *geojson.FeatureCollection) *geojson.FeatureCollection {
	result := make([]*geojson.Feature, 0, len(fc.Features))
	for _, f := range fc.Features {
		if f == nil || f.Geometry == nil {
			continue
		}
		if isEmptyGeometry(f.Geometry) || hasNaNCoords(f.Geometry) {
			continue
		}
		result = append(result, f)
	}
	return &geojson.FeatureCollection{Type: "FeatureCollection", Features: result}
}

// CountDuplicateCoordinates returns the number of duplicate coordinate groups.
func CountDuplicateCoordinates(fc *geojson.FeatureCollection) int {
	idx := spatial.NewIndex(fc)
	dupes := idx.FindDuplicateCoordinates(6)
	return len(dupes)
}

// AveragePropertyCount returns average number of properties per feature.
func AveragePropertyCount(fc *geojson.FeatureCollection) float64 {
	if len(fc.Features) == 0 {
		return 0
	}
	total := 0
	for _, f := range fc.Features {
		if f.Properties != nil {
			total += len(f.Properties)
		}
	}
	return float64(total) / float64(len(fc.Features))
}

// DominantGeometryType returns the most common geometry type.
func DominantGeometryType(fc *geojson.FeatureCollection) string {
	counts := make(map[string]int)
	for _, f := range fc.Features {
		if f != nil && f.Geometry != nil {
			counts[f.Geometry.GeoJSONType()]++
		}
	}
	best := ""
	bestCount := 0
	for t, c := range counts {
		if c > bestCount {
			best = t
			bestCount = c
		}
	}
	return best
}

// FormatReport returns human-readable validation output.
func FormatReport(r Report) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Valid: %v\n", r.Valid)
	fmt.Fprintf(&b, "Features: %d (valid: %d, invalid: %d)\n", r.FeatureCount, r.ValidFeatures, r.InvalidFeatures)
	if len(r.Issues) > 0 {
		b.WriteString("\nIssues:\n")
		limit := len(r.Issues)
		if limit > 20 {
			limit = 20
		}
		for _, issue := range r.Issues[:limit] {
			fmt.Fprintf(&b, "  [%s] %s", issue.Severity, issue.Code)
			if issue.Index >= 0 {
				fmt.Fprintf(&b, " (feature %d)", issue.Index)
			}
			fmt.Fprintf(&b, ": %s\n", issue.Message)
		}
		if len(r.Issues) > 20 {
			fmt.Fprintf(&b, "  ... and %d more issues\n", len(r.Issues)-20)
		}
	}
	return b.String()
}
