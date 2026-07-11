package pipeline

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/paulmach/orb"

	"github.com/BaconFries/meshtastic-poi/internal/model"
	"github.com/BaconFries/meshtastic-poi/internal/spatial"
)

// Validate checks POIs and drops invalid entries unless ValidateOnly is set via context.
type Validate struct{}

func (Validate) Name() string { return "validate" }

func (Validate) Process(_ context.Context, pois []*model.POI) ([]*model.POI, error) {
	out := make([]*model.POI, 0, len(pois))
	for _, p := range pois {
		if p != nil && p.Valid() && validLocation(p.Location) {
			out = append(out, p)
		}
	}
	return out, nil
}

// RemoveInvalid drops POIs with unusable coordinates.
type RemoveInvalid struct{}

func (RemoveInvalid) Name() string { return "remove_invalid" }

func (RemoveInvalid) Process(_ context.Context, pois []*model.POI) ([]*model.POI, error) {
	return Validate{}.Process(context.Background(), pois)
}

// Normalize rounds coordinates to the configured precision.
type Normalize struct {
	Precision int
}

func (n Normalize) Name() string { return "normalize" }

func (n Normalize) Process(_ context.Context, pois []*model.POI) ([]*model.POI, error) {
	precision := n.Precision
	if precision <= 0 {
		precision = 7
	}
	for _, p := range pois {
		if p == nil {
			continue
		}
		p.Location = orb.Point{
			round(p.Location[0], precision),
			round(p.Location[1], precision),
		}
	}
	return pois, nil
}

// RemoveEmpty strips empty tag and metadata values.
type RemoveEmpty struct{}

func (RemoveEmpty) Name() string { return "remove_empty" }

func (RemoveEmpty) Process(_ context.Context, pois []*model.POI) ([]*model.POI, error) {
	for _, p := range pois {
		if p == nil {
			continue
		}
		for k, v := range p.Tags {
			if strings.TrimSpace(v) == "" {
				delete(p.Tags, k)
			}
		}
		for k, v := range p.Metadata {
			if v == nil {
				delete(p.Metadata, k)
				continue
			}
			if s, ok := v.(string); ok && strings.TrimSpace(s) == "" {
				delete(p.Metadata, k)
			}
		}
	}
	return pois, nil
}

// Compress keeps only essential POI fields in tags.
type Compress struct{}

func (Compress) Name() string { return "compress" }

func (c Compress) Process(ctx context.Context, pois []*model.POI) ([]*model.POI, error) {
	return Simplify{Keep: []string{"name", "title", "label", "type", "id", "description"}}.Process(ctx, pois)
}

// Simplify drops or keeps tag/metadata fields.
type Simplify struct {
	Drop []string
	Keep []string
}

func (Simplify) Name() string { return "simplify" }

func (s Simplify) Process(_ context.Context, pois []*model.POI) ([]*model.POI, error) {
	drop := toSet(s.Drop)
	keep := toSet(s.Keep)
	useKeep := len(keep) > 0
	for _, p := range pois {
		if p == nil {
			continue
		}
		for k := range p.Tags {
			if containsFold(drop, k) || (useKeep && !containsFold(keep, k)) {
				delete(p.Tags, k)
			}
		}
		for k := range p.Metadata {
			if k == "geometry" || k == "geometry_type" {
				continue
			}
			if containsFold(drop, k) || (useKeep && !containsFold(keep, k)) {
				delete(p.Metadata, k)
			}
		}
	}
	return pois, nil
}

// Deduplicate removes duplicate POIs by ID and coordinates.
type Deduplicate struct {
	Precision int
}

func (Deduplicate) Name() string { return "deduplicate" }

func (d Deduplicate) Process(_ context.Context, pois []*model.POI) ([]*model.POI, error) {
	precision := d.Precision
	if precision <= 0 {
		precision = 6
	}
	seenID := make(map[string]struct{})
	seenCoord := make(map[string]struct{})
	out := make([]*model.POI, 0, len(pois))
	for _, p := range pois {
		if p == nil {
			continue
		}
		if p.ID != "" {
			if _, ok := seenID[p.ID]; ok {
				continue
			}
			seenID[p.ID] = struct{}{}
		}
		key := coordKey(p.Location, precision)
		if key != "" {
			if _, ok := seenCoord[key]; ok {
				continue
			}
			seenCoord[key] = struct{}{}
		}
		out = append(out, p)
	}
	return out, nil
}

// SortByID sorts POIs by ID then name.
type SortByID struct{}

func (SortByID) Name() string { return "sort_by_id" }

func (SortByID) Process(_ context.Context, pois []*model.POI) ([]*model.POI, error) {
	sort.SliceStable(pois, func(i, j int) bool {
		return sortKey(pois[i]) < sortKey(pois[j])
	})
	return pois, nil
}

// SortByDistance sorts POIs by distance from a reference point.
type SortByDistance struct {
	Lat float64
	Lon float64
}

func (SortByDistance) Name() string { return "sort_by_distance" }

func (s SortByDistance) Process(_ context.Context, pois []*model.POI) ([]*model.POI, error) {
	ref := spatial.DistanceFrom{Lat: s.Lat, Lon: s.Lon}
	type item struct {
		p    *model.POI
		dist float64
	}
	items := make([]item, 0, len(pois))
	for _, p := range pois {
		if p == nil {
			continue
		}
		items = append(items, item{p: p, dist: ref.DistanceTo(p.Location)})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].dist < items[j].dist
	})
	out := make([]*model.POI, len(items))
	for i, it := range items {
		out[i] = it.p
	}
	return out, nil
}

func sortKey(p *model.POI) string {
	if p == nil {
		return ""
	}
	if p.ID != "" {
		return p.ID
	}
	if p.Name != "" {
		return p.Name
	}
	return coordKey(p.Location, 6)
}

func coordKey(p orb.Point, precision int) string {
	factor := math.Pow(10, float64(precision))
	lon := math.Round(p[0]*factor) / factor
	lat := math.Round(p[1]*factor) / factor
	return fmt.Sprintf("%.6f,%.6f", lon, lat)
}

func round(v float64, precision int) float64 {
	factor := math.Pow(10, float64(precision))
	return math.Round(v*factor) / factor
}

func validLocation(p orb.Point) bool {
	if math.IsNaN(p[0]) || math.IsNaN(p[1]) || math.IsInf(p[0], 0) || math.IsInf(p[1], 0) {
		return false
	}
	return p[0] >= -180 && p[0] <= 180 && p[1] >= -90 && p[1] <= 90
}

func toSet(fields []string) map[string]struct{} {
	s := make(map[string]struct{}, len(fields))
	for _, f := range fields {
		s[strings.ToLower(f)] = struct{}{}
	}
	return s
}

func containsFold(set map[string]struct{}, key string) bool {
	_, ok := set[strings.ToLower(key)]
	return ok
}
