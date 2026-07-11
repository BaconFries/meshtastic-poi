package cli

import (
	"fmt"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/BaconFries/meshtastic-poi/internal/exporters"
	"github.com/BaconFries/meshtastic-poi/internal/model"
	"github.com/BaconFries/meshtastic-poi/internal/optimizer"
)

var (
	splitByField     string
	splitMaxFeatures int
	splitTileSize    float64
	splitOutputDir   string
)

var splitCmd = &cobra.Command{
	Use:   "split [file]",
	Short: "Split a dataset by field, tile, or feature count",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		setupLogging(verbose)
		pois, err := loadPOIs(args[0])
		if err != nil {
			log.Fatal().Err(err).Msg("read input")
		}

		fc := model.ToFeatureCollection(pois)
		opts := optimizer.SplitOptions{
			ByField:     splitByField,
			MaxFeatures: splitMaxFeatures,
			TileSizeDeg: splitTileSize,
		}
		groups, err := optimizer.Split(fc, opts)
		if err != nil {
			log.Fatal().Err(err).Msg("split failed")
		}

		outDir := splitOutputDir
		if outDir == "" {
			outDir = "."
		}

		for name, group := range groups {
			out := filepath.Join(outDir, fmt.Sprintf("%s.geojson", name))
			if outputPath != "" && len(groups) == 1 {
				out = outputPath
			}
			groupPOIs := model.FromFeatureCollection(group, "split")
			if err := exporters.WriteGeoJSONFile(out, groupPOIs); err != nil {
				log.Fatal().Err(err).Str("output", out).Msg("write failed")
			}
			log.Info().Str("output", out).Int("pois", len(groupPOIs)).Msg("wrote split")
		}
	},
}

func init() {
	splitCmd.Flags().StringVar(&splitByField, "by", "", "split by field (county, park, state, etc.)")
	splitCmd.Flags().IntVar(&splitMaxFeatures, "max-features", 0, "split by max features per file")
	splitCmd.Flags().Float64Var(&splitTileSize, "tile-size", 0, "split by tile size in degrees")
	splitCmd.Flags().StringVar(&splitOutputDir, "output-dir", "", "directory for split output files")
}
