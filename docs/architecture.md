# Architecture

## Overview

`meshtastic-poi` is structured as a reusable GIS toolkit with clear separation between data acquisition, processing, spatial operations, and output.

```
cmd/meshtastic-poi/     CLI entrypoint (Cobra)
internal/
  cli/                  Command handlers
  config/               Viper YAML configuration
  providers/            Provider interface + registry
    arcgis/             ArcGIS FeatureServer implementation
    geojson/            GeoJSON URL provider
    csv/                CSV point provider
    osm/                Overpass API provider
    gpx/                GPX file/URL provider
    kml/                KML file/URL provider
    garmin/             Garmin POI CSV provider
    source/             Local file and URL reader
    register/           Provider registration (avoids import cycles)
  downloader/           HTTP client with retries
  optimizer/            Validation + optimization pipeline
  spatial/              Quadtree index, filters, split helpers
  output/               GeoJSON and Meshtastic writers
pkg/geo/                Public geometry helpers
```

## Data Flow

```
Config (YAML)
    ↓
Provider.Download()
    ↓
geojson.FeatureCollection
    ↓
Validate → Optimize → Filter → Split/Merge
    ↓
Output (GeoJSON / Meshtastic)
```

## Provider Interface

```go
type Provider interface {
    Name() string
    Download(ctx context.Context) (*geojson.FeatureCollection, error)
    Metadata(ctx context.Context) (*Metadata, error)
}
```

Providers are registered via factory functions in `internal/providers/register`. This keeps the core `providers` package free of implementation imports.

## Spatial Index

The spatial package uses `github.com/paulmach/orb/quadtree` for:

- Radius search
- Bounding box queries
- Nearest neighbor
- Duplicate coordinate detection

Features implement `orb.Pointer` via a thin wrapper.

## Optimization Pipeline

```
Read → Validate → Remove Invalid → Normalize → Remove Empty Properties
     → Simplify Properties → Deduplicate → Sort → Write
```

## Design Principles

1. **Provider-agnostic** — No Florida or Meshtastic assumptions in core logic
2. **No CGO** — Pure Go for cross-platform builds
3. **Linear memory** — Operates on feature slices, avoids unnecessary copies
4. **Extensible** — New providers and output formats via registration
