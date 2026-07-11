// Package model defines the canonical POI representation used throughout the engine.
// Providers convert external data into POI values; exporters convert POI values into
// target formats. No stage in the pipeline depends on GeoJSON or Meshtastic.
package model

import (
	"time"

	"github.com/paulmach/orb"
)

// POI is the canonical point-of-interest record for the engine.
type POI struct {
	ID          string
	Name        string
	Category    string
	Description string

	Location orb.Point

	Address string

	Tags     map[string]string
	Metadata map[string]any

	Source  string
	Updated time.Time
}

// Clone returns a deep copy of the POI.
func (p *POI) Clone() *POI {
	if p == nil {
		return nil
	}
	cp := *p
	if p.Tags != nil {
		cp.Tags = make(map[string]string, len(p.Tags))
		for k, v := range p.Tags {
			cp.Tags[k] = v
		}
	}
	if p.Metadata != nil {
		cp.Metadata = make(map[string]any, len(p.Metadata))
		for k, v := range p.Metadata {
			cp.Metadata[k] = v
		}
	}
	return &cp
}

// Valid reports whether the POI has usable coordinates.
func (p *POI) Valid() bool {
	if p == nil {
		return false
	}
	lon, lat := p.Location[0], p.Location[1]
	if lon < -180 || lon > 180 || lat < -90 || lat > 90 {
		return false
	}
	return !(lon == 0 && lat == 0 && p.Name == "" && p.ID == "")
}

// Merge combines POI slices preserving order.
func Merge(slices ...[]*POI) []*POI {
	total := 0
	for _, s := range slices {
		total += len(s)
	}
	out := make([]*POI, 0, total)
	for _, s := range slices {
		out = append(out, s...)
	}
	return out
}
