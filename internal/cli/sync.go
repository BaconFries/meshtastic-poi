package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/paulmach/orb/geojson"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/BaconFries/meshtastic-poi/internal/config"
	"github.com/BaconFries/meshtastic-poi/internal/optimizer"
	"github.com/BaconFries/meshtastic-poi/internal/output"
	"github.com/BaconFries/meshtastic-poi/internal/providers"
	"github.com/BaconFries/meshtastic-poi/internal/providers/register"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Download, validate, optimize, and merge all configured sources",
	Run: func(cmd *cobra.Command, args []string) {
		setupLogging(verbose)
		if configPath == "" {
			log.Fatal().Msg("--config is required for sync")
		}
		cfg, err := config.Load(configPath)
		exitOnError(err)

		registry := register.DefaultRegistry()
		deps := providers.Dependencies{CacheDir: cfg.CacheDir}
		ctx := context.Background()

		collections := make([]*geojson.FeatureCollection, 0, len(cfg.Sources))
		for _, src := range cfg.Sources {
			p, err := registry.Create(src, deps)
			if err != nil {
				log.Fatal().Err(err).Str("source", src.Name).Msg("create provider")
			}
			log.Info().Str("source", p.Name()).Msg("downloading")
			fc, err := p.Download(ctx)
			if err != nil {
				log.Fatal().Err(err).Str("source", src.Name).Msg("download failed")
			}

			report := optimizer.Validate(fc)
			if !report.Valid {
				log.Warn().Str("source", src.Name).Int("invalid", report.InvalidFeatures).Msg("validation issues found")
			}

			fc = optimizer.Pipeline(fc, optimizer.Options{
				Minimal:     true,
				Dedupe:      true,
				RemoveEmpty: true,
			})
			collections = append(collections, fc)
		}

		merged := output.Merge(collections...)
		outDir := cfg.Output.Dir
		if outDir == "" {
			outDir = "."
		}
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			log.Fatal().Err(err).Msg("create output dir")
		}

		filename := cfg.Output.Filename
		if filename == "" {
			filename = "combined.geojson"
		}
		outPath := outputPath
		if outPath == "" {
			outPath = filepath.Join(outDir, filename)
		}

		format := cfg.Output.Format
		if format == "" {
			format = "geojson"
		}
		outputFormat = format
		if err := writeOutput(outPath, merged); err != nil {
			log.Fatal().Err(err).Msg("write output")
		}
		log.Info().Str("output", outPath).Int("features", len(merged.Features)).Msg("sync complete")
		fmt.Printf("Wrote %d features to %s\n", len(merged.Features), outPath)
	},
}
