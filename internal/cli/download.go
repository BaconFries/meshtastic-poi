package cli

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/BaconFries/meshtastic-poi/internal/config"
	"github.com/BaconFries/meshtastic-poi/internal/output"
	"github.com/BaconFries/meshtastic-poi/internal/providers"
	"github.com/BaconFries/meshtastic-poi/internal/providers/register"
)

var downloadSource string

var downloadCmd = &cobra.Command{
	Use:   "download [source-name]",
	Short: "Download POI data from configured sources",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		setupLogging(verbose)
		cfg, err := config.Load(configPath)
		exitOnError(err)

		if len(args) > 0 {
			downloadSource = args[0]
		}

		registry := register.DefaultRegistry()
		deps := providers.Dependencies{CacheDir: cfg.CacheDir}

		sources := cfg.Sources
		if downloadSource != "" {
			sources = filterSources(cfg.Sources, downloadSource)
			if len(sources) == 0 {
				log.Fatal().Str("source", downloadSource).Msg("source not found in config")
			}
		}

		ctx := context.Background()
		for _, src := range sources {
			p, err := registry.Create(src, deps)
			if err != nil {
				log.Fatal().Err(err).Str("source", src.Name).Msg("create provider")
			}
			log.Info().Str("source", p.Name()).Msg("downloading")
			fc, err := p.Download(ctx)
			if err != nil {
				log.Fatal().Err(err).Str("source", src.Name).Msg("download failed")
			}
			out := outputPath
			if out == "" {
				out = fmt.Sprintf("%s.geojson", sanitize(src.Name))
			}
			if err := output.WriteGeoJSON(out, fc); err != nil {
				log.Fatal().Err(err).Str("output", out).Msg("write failed")
			}
			log.Info().Str("output", out).Int("features", len(fc.Features)).Msg("download complete")
		}
	},
}

func filterSources(sources []providers.SourceConfig, name string) []providers.SourceConfig {
	var result []providers.SourceConfig
	for _, s := range sources {
		if s.Name == name {
			result = append(result, s)
		}
	}
	return result
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
		return "output"
	}
	return string(out)
}
