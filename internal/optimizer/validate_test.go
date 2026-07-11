package optimizer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BaconFries/meshtastic-poi/internal/optimizer"
	"github.com/BaconFries/meshtastic-poi/internal/output"
)

func TestValidateDetectsIssues(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "sample_pois.geojson"))
	if err != nil {
		t.Fatal(err)
	}
	report, err := optimizer.ValidateJSON(data)
	if err != nil {
		t.Fatal(err)
	}
	if report.FeatureCount != 5 {
		t.Fatalf("expected 5 features, got %d", report.FeatureCount)
	}
	if report.Valid {
		t.Fatal("expected invalid report due to null geometry")
	}
	if report.InvalidFeatures < 1 {
		t.Fatal("expected at least one invalid feature")
	}
	foundDupe := false
	for _, issue := range report.Issues {
		if issue.Code == "duplicate_coordinates" {
			foundDupe = true
		}
	}
	if !foundDupe {
		t.Fatal("expected duplicate coordinate warning")
	}
}

func TestPipelineDedupeAndNormalize(t *testing.T) {
	fc, err := output.ReadGeoJSON(filepath.Join("..", "..", "testdata", "dedupe_test.geojson"))
	if err != nil {
		t.Fatal(err)
	}
	result := optimizer.Pipeline(fc, optimizer.Options{
		Dedupe:     true,
		DropFields: []string{"OBJECTID", "Shape_Length"},
	})
	if len(result.Features) != 2 {
		t.Fatalf("expected 2 features after dedupe, got %d", len(result.Features))
	}
	for _, f := range result.Features {
		if f.Properties["OBJECTID"] != nil {
			t.Fatal("OBJECTID should be dropped")
		}
	}
}

func TestRemoveInvalidFeatures(t *testing.T) {
	fc, err := output.ReadGeoJSON(filepath.Join("..", "..", "testdata", "sample_pois.geojson"))
	if err != nil {
		t.Fatal(err)
	}
	clean := optimizer.RemoveInvalidFeatures(fc)
	if len(clean.Features) != 4 {
		t.Fatalf("expected 4 valid features, got %d", len(clean.Features))
	}
}
