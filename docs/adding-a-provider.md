# Adding a Provider

Providers fetch external data and convert it into the canonical `model.POI` type.

## Steps

1. Create a package under `internal/providers/<name>/`
2. Implement the `Provider` interface from `internal/providers`
3. Register the factory in `internal/providers/register/register.go`

## Interface

```go
type Provider interface {
    Name() string
    Metadata(context.Context) (*DatasetInfo, error)
    Fetch(context.Context) ([]*model.POI, error)
}
```

## Example Skeleton

```go
package myprovider

import (
    "context"
    "github.com/BaconFries/meshtastic-poi/internal/model"
    "github.com/BaconFries/meshtastic-poi/internal/providers"
)

type Provider struct {
    cfg providers.SourceConfig
}

func New(cfg providers.SourceConfig, deps providers.Dependencies) (providers.Provider, error) {
    return &Provider{cfg: cfg}, nil
}

func (p *Provider) Name() string { return p.cfg.Name }

func (p *Provider) Fetch(ctx context.Context) ([]*model.POI, error) {
    // Fetch external data and convert to []*model.POI
    return nil, nil
}

func (p *Provider) Metadata(ctx context.Context) (*providers.DatasetInfo, error) {
    return &providers.DatasetInfo{Name: p.Name(), Type: "myprovider"}, nil
}
```

## Registration

```go
r.Register("myprovider", myprovider.New)
```

## Configuration

```yaml
sources:
  - name: My Data
    type: myprovider
    url: https://example.com/data
    params:
      custom_option: value
```

## Testing

Add unit tests with `httptest` mock servers. See `internal/providers/arcgis/arcgis_test.go`.
