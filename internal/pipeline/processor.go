// Package pipeline implements composable POI processing stages.
package pipeline

import (
	"context"
	"fmt"

	"github.com/BaconFries/meshtastic-poi/internal/model"
)

// Processor transforms a slice of POIs. Stages must not depend on data origin or export format.
type Processor interface {
	Name() string
	Process(ctx context.Context, pois []*model.POI) ([]*model.POI, error)
}

// Options configures the default processing pipeline.
type Options struct {
	Minimal            bool
	RemoveEmpty        bool
	Dedupe             bool
	SortDistance       bool
	CompressProperties bool
	DropFields         []string
	KeepFields         []string
	SortLat            float64
	SortLon            float64
	CoordPrecision     int
	ValidateOnly       bool
}

// Run executes processors in order.
func Run(ctx context.Context, pois []*model.POI, processors []Processor) ([]*model.POI, error) {
	out := pois
	for _, p := range processors {
		if p == nil {
			continue
		}
		next, err := p.Process(ctx, out)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", p.Name(), err)
		}
		out = next
	}
	return out, nil
}

// Default returns the standard optimization pipeline for the given options.
func Default(opts Options) []Processor {
	stages := []Processor{
		RemoveInvalid{},
		Normalize{Precision: opts.CoordPrecision},
	}
	if opts.RemoveEmpty || opts.Minimal {
		stages = append(stages, RemoveEmpty{})
	}
	if opts.CompressProperties || opts.Minimal {
		stages = append(stages, Compress{})
	}
	if len(opts.DropFields) > 0 || len(opts.KeepFields) > 0 {
		stages = append(stages, Simplify{Drop: opts.DropFields, Keep: opts.KeepFields})
	}
	if opts.Dedupe || opts.Minimal {
		stages = append(stages, Deduplicate{Precision: opts.CoordPrecision})
	}
	if opts.SortDistance && opts.SortLat != 0 && opts.SortLon != 0 {
		stages = append(stages, SortByDistance{Lat: opts.SortLat, Lon: opts.SortLon})
	} else {
		stages = append(stages, SortByID{})
	}
	return stages
}

// ValidatePipeline returns validation-only stages.
func ValidatePipeline() []Processor {
	return []Processor{Validate{}}
}
