package cli

import (
	"github.com/paulmach/orb/geojson"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/BaconFries/meshtastic-poi/internal/output"
)

var mergeCmd = &cobra.Command{
	Use:   "merge [files...]",
	Short: "Merge multiple GeoJSON datasets into one",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		setupLogging(verbose)
		collections := make([]*geojson.FeatureCollection, 0, len(args))
		for _, path := range args {
			fc, err := output.ReadGeoJSON(path)
			if err != nil {
				log.Fatal().Err(err).Str("file", path).Msg("read input")
			}
			collections = append(collections, fc)
		}

		merged := output.Merge(collections...)
		out := outputPath
		if out == "" {
			out = "-"
		}
		if err := writeOutput(out, merged); err != nil {
			log.Fatal().Err(err).Msg("write output")
		}
		log.Info().Int("features", len(merged.Features)).Msg("merge complete")
	},
}
