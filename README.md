# meshtastic-poi

Cross-platform command-line GIS toolkit for downloading, validating, optimizing, filtering, and exporting Points of Interest (POI) data. Built for Meshtastic workflows but designed as a reusable, provider-agnostic geospatial tool.

## Features

- **Pluggable providers**: ArcGIS, GeoJSON URL, CSV, OSM (Overpass), GPX, KML, Garmin POI CSV
- **ArcGIS automation**: Auto-detects `maxRecordCount`, paginates all records, retries failures
- **Validation**: Invalid geometry, duplicates, malformed properties — console + JSON reports
- **Optimization pipeline**: Normalize, dedupe, simplify properties, sort
- **Spatial filtering**: Radius, bounding box, attribute filters with quadtree index
- **Split & merge**: By county, park, tile, or feature count
- **Sync workflow**: Download → validate → optimize → merge from config
- **Output formats**: GeoJSON, Meshtastic POI JSON

## Requirements

- Go 1.25+

## Install

```bash
go install ./cmd/meshtastic-poi
```

Or build locally:

```bash
go build -o bin/meshtastic-poi ./cmd/meshtastic-poi
```

## Quick Start

```bash
# List providers
meshtastic-poi providers

# Validate a dataset
meshtastic-poi validate testdata/sample_pois.geojson

# Show statistics
meshtastic-poi stats testdata/sample_pois.geojson

# Optimize with deduplication
meshtastic-poi optimize testdata/dedupe_test.geojson --dedupe --drop-fields OBJECTID,Shape_Length -o clean.geojson

# Filter by radius (meters)
meshtastic-poi filter testdata/sample_pois.geojson --lat 28.5383 --lon -81.3792 --radius 50000

# Merge datasets
meshtastic-poi merge parks.geojson campgrounds.geojson -o combined.geojson

# Full sync from config
meshtastic-poi sync --config examples/config.yaml
```

## CLI Commands

| Command | Description |
|---------|-------------|
| `download` | Download from configured sources |
| `validate` | Validate GeoJSON |
| `stats` | Dataset statistics |
| `optimize` | Run optimization pipeline |
| `filter` | Spatial and attribute filtering |
| `split` | Split by field, tile, or count |
| `merge` | Combine multiple datasets |
| `sync` | Full pipeline from config |
| `providers` | List available providers |

All commands support `--config`, `--verbose`, and `--output`.

## Configuration

See `examples/config.yaml` for a full example. Sources are defined with `name`, `type`, and `url`.

## Documentation

- [Architecture](docs/architecture.md)
- [Adding a Provider](docs/adding-a-provider.md)
- [Adding an Output Format](docs/adding-an-output-format.md)
- [Performance Notes](docs/performance.md)

## License

MIT
