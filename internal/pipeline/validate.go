package pipeline

import (
	"fmt"
	"strings"

	"github.com/BaconFries/meshtastic-poi/internal/model"
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

// Report aggregates validation results for a POI dataset.
type Report struct {
	Valid       bool    `json:"valid"`
	POICount    int     `json:"poi_count"`
	ValidPOIs   int     `json:"valid_pois"`
	InvalidPOIs int     `json:"invalid_pois"`
	Issues      []Issue `json:"issues"`
}

// ValidateReport checks POIs and returns a detailed report without modifying input.
func ValidateReport(pois []*model.POI) Report {
	report := Report{
		Valid:    true,
		POICount: len(pois),
		Issues:   make([]Issue, 0),
	}
	seenIDs := make(map[string]int)
	seenCoord := make(map[string]int)

	for i, p := range pois {
		if p == nil {
			report.add(SeverityError, "null_poi", "POI is nil", i, "")
			continue
		}
		if !validLocation(p.Location) {
			report.add(SeverityError, "invalid_location", "location is out of range or NaN", i, "location")
			continue
		}
		if p.ID != "" {
			if prev, ok := seenIDs[p.ID]; ok {
				report.add(SeverityWarning, "duplicate_id",
					fmt.Sprintf("duplicate id %q (also at index %d)", p.ID, prev), i, "id")
			}
			seenIDs[p.ID] = i
		}
		key := coordKey(p.Location, 6)
		if key != "" {
			if prev, ok := seenCoord[key]; ok {
				report.add(SeverityWarning, "duplicate_coordinates",
					fmt.Sprintf("duplicate coordinates (also at index %d)", prev), i, "location")
			}
			seenCoord[key] = i
		}
		report.ValidPOIs++
	}
	report.InvalidPOIs = report.POICount - report.ValidPOIs
	if report.InvalidPOIs > 0 {
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

// FormatReport returns human-readable validation output.
func FormatReport(r Report) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Valid: %v\n", r.Valid)
	fmt.Fprintf(&b, "POIs: %d (valid: %d, invalid: %d)\n", r.POICount, r.ValidPOIs, r.InvalidPOIs)
	if len(r.Issues) > 0 {
		b.WriteString("\nIssues:\n")
		limit := len(r.Issues)
		if limit > 20 {
			limit = 20
		}
		for _, issue := range r.Issues[:limit] {
			fmt.Fprintf(&b, "  [%s] %s", issue.Severity, issue.Code)
			if issue.Index >= 0 {
				fmt.Fprintf(&b, " (poi %d)", issue.Index)
			}
			fmt.Fprintf(&b, ": %s\n", issue.Message)
		}
		if len(r.Issues) > 20 {
			fmt.Fprintf(&b, "  ... and %d more issues\n", len(r.Issues)-20)
		}
	}
	return b.String()
}
