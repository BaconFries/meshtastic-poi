package pack

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/BaconFries/meshtastic-poi/internal/catalog"
	"github.com/BaconFries/meshtastic-poi/internal/config"
	"github.com/BaconFries/meshtastic-poi/internal/exporters"
	"github.com/BaconFries/meshtastic-poi/internal/model"
	"github.com/BaconFries/meshtastic-poi/internal/pipeline"
	"github.com/BaconFries/meshtastic-poi/internal/providers"
	"github.com/BaconFries/meshtastic-poi/internal/providers/register"
)

// Definition describes a POI pack collection.
type Definition struct {
	Name      string           `yaml:"name" json:"name"`
	Sources   []string         `yaml:"sources" json:"sources"`
	Optimizer pipeline.Options `yaml:"optimizer" json:"optimizer"`
	Exports   []string         `yaml:"exports" json:"exports"`
}

// Load reads a pack definition from YAML.
func Load(path string) (*Definition, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var def Definition
	if err := yaml.Unmarshal(data, &def); err != nil {
		return nil, fmt.Errorf("parse pack: %w", err)
	}
	return &def, nil
}

// Build fetches catalog datasets, processes them, and exports configured formats.
func Build(ctx context.Context, def *Definition, cat *catalog.Catalog, cfg config.Config, outDir string) ([]*model.POI, error) {
	if def == nil {
		return nil, fmt.Errorf("pack definition is nil")
	}
	registry := register.DefaultRegistry()
	deps := providers.Dependencies{CacheDir: cfg.CacheDir}

	var all []*model.POI
	for _, sourceID := range def.Sources {
		ds, ok := cat.Get(sourceID)
		if !ok {
			return nil, fmt.Errorf("catalog dataset %q not found", sourceID)
		}
		src := providers.SourceConfig{
			Name: ds.Name,
			Type: ds.Provider,
			URL:  ds.URL,
		}
		p, err := registry.Create(src, deps)
		if err != nil {
			return nil, err
		}
		pois, err := p.Fetch(ctx)
		if err != nil {
			return nil, fmt.Errorf("fetch %s: %w", sourceID, err)
		}
		all = append(all, pois...)
	}

	processed, err := pipeline.Run(ctx, all, pipeline.Default(def.Optimizer))
	if err != nil {
		return nil, err
	}

	if outDir == "" {
		outDir = "."
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, err
	}
	for _, format := range def.Exports {
		path := filepath.Join(outDir, sanitize(def.Name)+"."+formatExt(format))
		if err := exporters.WriteFile(path, format, processed); err != nil {
			return nil, fmt.Errorf("export %s: %w", format, err)
		}
	}
	return processed, nil
}

func formatExt(format string) string {
	switch format {
	case "meshtastic":
		return "json"
	default:
		return format
	}
}

func sanitize(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' {
			out = append(out, c)
		} else if c == ' ' {
			out = append(out, '_')
		}
	}
	if len(out) == 0 {
		return "pack"
	}
	return string(out)
}
