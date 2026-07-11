package gpx

import (
	"context"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"

	"github.com/BaconFries/meshtastic-poi/internal/downloader"
	"github.com/BaconFries/meshtastic-poi/internal/model"
	"github.com/BaconFries/meshtastic-poi/internal/providers"
	"github.com/BaconFries/meshtastic-poi/internal/providers/source"
)

// Provider reads GPX files from URL or local path.
type Provider struct {
	cfg    providers.SourceConfig
	client *downloader.Client
}

type gpxDoc struct {
	XMLName xml.Name `xml:"gpx"`
	Waypts  []gpxWpt `xml:"wpt"`
	Routes  []gpxRte `xml:"rte"`
	Tracks  []gpxTrk `xml:"trk"`
}

type gpxWpt struct {
	Lat  float64 `xml:"lat,attr"`
	Lon  float64 `xml:"lon,attr"`
	Name string  `xml:"name"`
	Desc string  `xml:"desc"`
	Type string  `xml:"type"`
}

type gpxRte struct {
	Name string     `xml:"name"`
	Desc string     `xml:"desc"`
	Pts  []gpxRtePt `xml:"rtept"`
}

type gpxRtePt struct {
	Lat  float64 `xml:"lat,attr"`
	Lon  float64 `xml:"lon,attr"`
	Name string  `xml:"name"`
}

type gpxTrk struct {
	Name string      `xml:"name"`
	Desc string      `xml:"desc"`
	Segs []gpxTrkSeg `xml:"trkseg"`
}

type gpxTrkSeg struct {
	Pts []gpxTrkPt `xml:"trkpt"`
}

type gpxTrkPt struct {
	Lat float64 `xml:"lat,attr"`
	Lon float64 `xml:"lon,attr"`
}

// New creates a GPX provider.
func New(cfg providers.SourceConfig, deps providers.Dependencies) (providers.Provider, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("gpx provider requires url")
	}
	return &Provider{
		cfg:    cfg,
		client: downloader.New(deps.CacheDir),
	}, nil
}

func (p *Provider) Name() string {
	if p.cfg.Name != "" {
		return p.cfg.Name
	}
	return "gpx"
}

func (p *Provider) Metadata(ctx context.Context) (*providers.DatasetInfo, error) {
	pois, err := p.Fetch(ctx)
	if err != nil {
		return nil, err
	}
	return &providers.DatasetInfo{
		Name:     p.Name(),
		Type:     "gpx",
		URL:      p.cfg.URL,
		POICount: len(pois),
	}, nil
}

func (p *Provider) Fetch(ctx context.Context) ([]*model.POI, error) {
	body, err := source.Read(ctx, p.client, p.cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("read gpx: %w", err)
	}
	fc, err := Parse(body)
	if err != nil {
		return nil, err
	}
	return model.FromFeatureCollection(fc, p.Name()), nil
}

// Parse converts GPX XML bytes to GeoJSON.
func Parse(data []byte) (*geojson.FeatureCollection, error) {
	var doc gpxDoc
	if err := xml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse gpx: %w", err)
	}

	fc := &geojson.FeatureCollection{
		Type:     "FeatureCollection",
		Features: make([]*geojson.Feature, 0),
	}

	for i, wpt := range doc.Waypts {
		f := geojson.NewFeature(orb.Point{wpt.Lon, wpt.Lat})
		f.Properties = geojson.Properties{
			"name": wpt.Name,
			"desc": wpt.Desc,
			"type": wpt.Type,
			"gpx":  "waypoint",
		}
		f.ID = "wpt/" + strconv.Itoa(i)
		fc.Features = append(fc.Features, f)
	}

	for i, rte := range doc.Routes {
		if len(rte.Pts) == 0 {
			continue
		}
		line := make(orb.LineString, len(rte.Pts))
		for j, pt := range rte.Pts {
			line[j] = orb.Point{pt.Lon, pt.Lat}
		}
		f := geojson.NewFeature(line)
		f.Properties = geojson.Properties{
			"name": rte.Name,
			"desc": rte.Desc,
			"gpx":  "route",
		}
		f.ID = "rte/" + strconv.Itoa(i)
		fc.Features = append(fc.Features, f)
	}

	for i, trk := range doc.Tracks {
		for j, seg := range trk.Segs {
			if len(seg.Pts) == 0 {
				continue
			}
			line := make(orb.LineString, len(seg.Pts))
			for k, pt := range seg.Pts {
				line[k] = orb.Point{pt.Lon, pt.Lat}
			}
			f := geojson.NewFeature(line)
			name := trk.Name
			if len(trk.Segs) > 1 {
				name = strings.TrimSpace(fmt.Sprintf("%s segment %d", trk.Name, j+1))
			}
			f.Properties = geojson.Properties{
				"name": name,
				"desc": trk.Desc,
				"gpx":  "track",
			}
			f.ID = fmt.Sprintf("trk/%d/%d", i, j)
			fc.Features = append(fc.Features, f)
		}
	}

	return fc, nil
}

func dominantType(fc *geojson.FeatureCollection) string {
	if len(fc.Features) == 0 {
		return ""
	}
	return fc.Features[0].Geometry.GeoJSONType()
}
