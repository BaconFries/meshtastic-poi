# CLI Reference

## Global Flags

| Flag | Description |
|------|-------------|
| `--config` | Config file path (YAML or JSON) |
| `--verbose` | Enable debug logging |
| `-o, --output` | Output file path |

## Data Commands

| Command | Description |
|---------|-------------|
| `download [source]` | Fetch POIs from configured sources |
| `sync` | Download, validate, optimize, and merge all sources |
| `validate [file]` | Validate a POI dataset |
| `stats [file]` | Show dataset statistics |
| `optimize [file]` | Run the optimization pipeline |
| `filter [file]` | Spatial and attribute filtering |
| `merge [files...]` | Combine multiple datasets |
| `split [file]` | Split by field, tile, or count |
| `export [file]` | Export to configured format (`--format`) |

## Catalog & Packs

| Command | Description |
|---------|-------------|
| `catalog list` | List catalog datasets |
| `catalog info [id]` | Show dataset details |
| `catalog add` | Add dataset (`--id`, `--name`, `--provider`, `--url`) |
| `catalog remove [id]` | Remove a dataset |
| `pack build [pack.yaml]` | Build a pack from catalog sources |

## Utilities

| Command | Description |
|---------|-------------|
| `providers` | List registered provider types |
| `index [file]` | Spatial index summary |
| `doctor` | Environment diagnostics |
| `benchmark [file]` | Performance metrics |

## Examples

```bash
# Optimize with deduplication and export as Meshtastic JSON
meshtastic-poi optimize data.geojson --dedupe --minimal --format meshtastic -o nodes.json

# Filter by radius
meshtastic-poi filter data.geojson --lat 28.5383 --lon -81.3792 --radius 50000

# Add a dataset to the catalog
meshtastic-poi catalog add --id florida-parks --name "FL Parks" --provider arcgis --url https://...

# Run diagnostics
meshtastic-poi doctor
```
