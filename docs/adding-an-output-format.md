# Adding an Exporter

Exporters write canonical POI datasets to external formats. They must not depend on provider types.

## Steps

1. Add an exporter in `internal/exporters/<format>.go`
2. Implement the `Exporter` interface
3. Register in `internal/exporters/register.go`
4. Wire into CLI via `--format` flag

## Interface

```go
type Exporter interface {
    Name() string
    Export(context.Context, []*model.POI, io.Writer) error
}
```

## Example

```go
type MyFormat struct{}

func (MyFormat) Name() string { return "myformat" }

func (MyFormat) Export(ctx context.Context, pois []*model.POI, w io.Writer) error {
    for _, p := range pois {
        // serialize p
    }
    return nil
}
```

## Meshtastic Format

The Meshtastic exporter converts POIs to:

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

## Built-in Exporters

| Name | Status |
|------|--------|
| `geojson` | Complete |
| `meshtastic` | Complete |
| `csv` | Complete |
| `gpx` | Stub (waypoints) |
| `kml` | Stub (points) |
| `garmin` | Stub (CSV) |
