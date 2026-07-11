package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/BaconFries/meshtastic-poi/internal/config"
	"github.com/BaconFries/meshtastic-poi/internal/model"
	"github.com/BaconFries/meshtastic-poi/internal/pipeline"
	"github.com/BaconFries/meshtastic-poi/internal/providers"
	"github.com/BaconFries/meshtastic-poi/internal/providers/register"
	"github.com/BaconFries/meshtastic-poi/pkg/engine"
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

		var all [][]*model.POI
		for _, src := range cfg.Sources {
			p, err := registry.Create(src, deps)
			if err != nil {
				log.Fatal().Err(err).Str("source", src.Name).Msg("create provider")
			}
			log.Info().Str("source", p.Name()).Msg("downloading")
			pois, err := p.Fetch(ctx)
			if err != nil {
				log.Fatal().Err(err).Str("source", src.Name).Msg("download failed")
			}

			report := pipeline.ValidateReport(pois)
			if !report.Valid {
				log.Warn().Str("source", src.Name).Int("invalid", report.InvalidPOIs).Msg("validation issues found")
			}

			processed, err := runPipeline(ctx, pois, pipeline.Options{
				Minimal:     true,
				Dedupe:      true,
				RemoveEmpty: true,
			})
			if err != nil {
				log.Fatal().Err(err).Str("source", src.Name).Msg("pipeline failed")
			}
			all = append(all, processed)
		}

		merged := engine.Merge(all...)
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
		if err := writePOIs(outPath, merged); err != nil {
			log.Fatal().Err(err).Msg("write output")
		}
		log.Info().Str("output", outPath).Int("pois", len(merged)).Msg("sync complete")
		fmt.Printf("Wrote %d POIs to %s\n", len(merged), outPath)
	},
}
