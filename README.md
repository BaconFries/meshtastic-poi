# meshtastic-poi

Cross-platform offline POI management engine. Meshtastic is the primary consumer today, but the core library is provider- and format-agnostic — everything flows through a canonical POI model with pluggable providers, processors, and exporters.

## Architecture

```
Providers → Canonical POI Model → Processing Pipeline → Exporters
```

See [Architecture](docs/architecture.md) for full details.

## Features

- **Canonical POI model** — Single internal representation for all data
- **Pluggable providers** — ArcGIS, GeoJSON, CSV, OSM, GPX, KML, Garmin
- **Composable pipeline** — Validate, normalize, dedupe, simplify, sort
- **Exporters** — GeoJSON, Meshtastic, CSV (+ GPX/KML/Garmin stubs)
- **Spatial engine** — Radius, bbox, nearest neighbor, duplicate detection
- **Catalog & packs** — Manage datasets and build collections
- **Embeddable API** — `pkg/engine` for use in other Go applications

## Requirements

- Go 1.25+

## Install

```bash
go install github.com/BaconFries/meshtastic-poi/cmd/meshtastic-poi@latest
```

Or build locally:

```bash
go build -o bin/meshtastic-poi ./cmd/meshtastic-poi
```

## Quick Start

```bash
meshtastic-poi providers
meshtastic-poi validate testdata/sample_pois.geojson
meshtastic-poi optimize testdata/dedupe_test.geojson --dedupe --minimal -o clean.geojson
meshtastic-poi export testdata/sample_pois.geojson --format meshtastic -o nodes.json
meshtastic-poi sync --config examples/config.yaml
meshtastic-poi doctor
```

## CLI Commands

| Command | Description |
|---------|-------------|
| `download` | Fetch POIs from configured sources |
| `sync` | Full pipeline from config |
| `validate` | Validate a dataset |
| `stats` | Dataset statistics |
| `optimize` | Run optimization pipeline |
| `filter` | Spatial and attribute filtering |
| `merge` | Combine datasets |
| `split` | Split by field, tile, or count |
| `export` | Export to target format |
| `catalog` | Manage dataset catalog |
| `pack` | Build POI packs |
| `doctor` | Environment diagnostics |
| `benchmark` | Performance metrics |

See [CLI Reference](docs/cli-reference.md).

## Documentation

- [Architecture](docs/architecture.md)
- [CLI Reference](docs/cli-reference.md)
- [Adding a Provider](docs/adding-a-provider.md)
- [Adding an Exporter](docs/adding-an-output-format.md)
- [Performance Notes](docs/performance.md)

## License

MIT
