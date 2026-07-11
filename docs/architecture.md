# Architecture

## Overview

`meshtastic-poi` is a provider-agnostic offline POI management engine. All data flows through a canonical internal model; no pipeline stage or exporter depends on where data originated.

```
               Providers
      ArcGIS  OSM  GPX  CSV  KML  Garmin
            │
            ▼
      Canonical POI Model  (internal/model)
            │
            ▼
        Processing Pipeline  (internal/pipeline)
            │
            ▼
 GeoJSON  Meshtastic  CSV  GPX  KML  Garmin
         (internal/exporters)
```

## Packages

```
cmd/meshtastic-poi/     CLI entrypoint (Cobra)
internal/
  model/                Canonical POI type + conversions
  pipeline/             Composable Processor stages
  providers/            Provider interface + implementations
  exporters/            Format exporters (GeoJSON, Meshtastic, CSV, …)
  spatial/              POI spatial index (quadtree), filters
  catalog/              Dataset catalog (~/.config/meshtastic-poi/catalog.yaml)
  pack/                 POI pack collections
  config/               Viper YAML/JSON configuration
  downloader/           HTTP client with retries
pkg/
  engine/               Public embeddable API
  geo/                  Geometry helpers
```

## Canonical POI Model

```go
type POI struct {
    ID, Name, Category, Description string
    Location orb.Point
    Address string
    Tags map[string]string
    Metadata map[string]any
    Source string
    Updated time.Time
}
```

Providers implement `Fetch(ctx) ([]*model.POI, error)`.
Exporters implement `Export(ctx, pois, io.Writer) error`.

## Processing Pipeline

Each stage implements:

```go
type Processor interface {
    Name() string
    Process(ctx context.Context, pois []*model.POI) ([]*model.POI, error)
}
```

Default order: remove invalid → normalize → remove empty → compress → dedupe → sort.

## Design Principles

1. **Provider-agnostic** — Core engine has no Meshtastic or GeoJSON dependencies
2. **No CGO** — Pure Go for cross-platform builds
3. **Composable** — Pipelines, catalogs, and packs assemble from interfaces
4. **Embeddable** — Use `pkg/engine` from other Go applications
5. **Extensible** — Register new providers and exporters without changing core logic

## CLI Commands

`download`, `sync`, `validate`, `stats`, `optimize`, `filter`, `merge`, `split`, `export`, `index`, `catalog`, `pack`, `doctor`, `benchmark`, `providers`

See [CLI Reference](cli-reference.md) for details.
