# Adding an Output Format

## Steps

1. Add a writer in `internal/output/<format>.go`
2. Wire the format into `internal/cli/optimize.go` `writeOutput()`
3. Optionally add config support in `internal/config/config.go`

## Example

```go
func WriteCSV(path string, fc *geojson.FeatureCollection) error {
    // Convert features to CSV rows
    return nil
}
```

## CLI Integration

Extend the `writeOutput` switch in `internal/cli/optimize.go`:

```go
case "csv":
    return output.WriteCSV(path, fc)
```

## Meshtastic Format

The Meshtastic exporter (`internal/output/meshtastic.go`) converts point features to:

```json
{
  "nodes": [
    {
      "name": "POI Name",
      "latitude": 28.5383,
      "longitude": -81.3792,
      "type": "park",
      "description": "Optional description"
    }
  ]
}
```

Non-point geometries use centroid coordinates.

## Future Formats

Planned: CSV, Garmin POI, GPX, KML. Each should live in `internal/output/` with a consistent `Write<Format>(path, fc)` signature.
