# Adding a Provider

## Steps

1. Create a package under `internal/providers/<name>/`
2. Implement the `Provider` interface from `internal/providers`
3. Register the factory in `internal/providers/register/register.go`

## Example Skeleton

```go
package myprovider

import (
    "context"
    "github.com/paulmach/orb/geojson"
    "github.com/BaconFries/meshtastic-poi/internal/providers"
)

type Provider struct {
    cfg providers.SourceConfig
}

func New(cfg providers.SourceConfig, deps providers.Dependencies) (providers.Provider, error) {
    return &Provider{cfg: cfg}, nil
}

func (p *Provider) Name() string { return p.cfg.Name }

func (p *Provider) Download(ctx context.Context) (*geojson.FeatureCollection, error) {
    // Fetch and convert to GeoJSON
    return nil, nil
}

func (p *Provider) Metadata(ctx context.Context) (*providers.Metadata, error) {
    return &providers.Metadata{Name: p.Name(), Type: "myprovider"}, nil
}
```

## Registration

```go
r.Register("myprovider", myprovider.New)
```

## Configuration

Users reference your provider in `config.yaml`:

```yaml
sources:
  - name: My Data
    type: myprovider
    url: https://example.com/data
    params:
      custom_option: value
```

## Shared Dependencies

Use `providers.Dependencies` for cache directory and future shared services (HTTP client, etc.).

## Testing

Add unit tests with `httptest` mock servers. See `internal/providers/arcgis/arcgis_test.go` for pagination examples.
