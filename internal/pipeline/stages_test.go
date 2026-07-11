package pipeline_test

import (
	"context"
	"testing"

	"github.com/paulmach/orb"

	"github.com/BaconFries/meshtastic-poi/internal/model"
	"github.com/BaconFries/meshtastic-poi/internal/pipeline"
)

func TestDeduplicate(t *testing.T) {
	pois := []*model.POI{
		{ID: "1", Location: orb.Point{-81, 28}, Name: "A"},
		{ID: "1", Location: orb.Point{-81, 28}, Name: "Dup"},
		{ID: "2", Location: orb.Point{-82, 29}, Name: "B"},
	}
	out, err := pipeline.Run(context.Background(), pois, []pipeline.Processor{pipeline.Deduplicate{}})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 POIs after dedupe, got %d", len(out))
	}
}

func TestValidateReport(t *testing.T) {
	pois := []*model.POI{
		{Location: orb.Point{-81, 28}, Name: "OK"},
		nil,
		{Location: orb.Point{999, 28}, Name: "Bad"},
	}
	report := pipeline.ValidateReport(pois)
	if report.Valid {
		t.Fatal("expected invalid report")
	}
	if report.ValidPOIs != 1 {
		t.Fatalf("expected 1 valid POI, got %d", report.ValidPOIs)
	}
}
