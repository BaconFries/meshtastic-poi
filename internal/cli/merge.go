package cli

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/BaconFries/meshtastic-poi/internal/model"
	"github.com/BaconFries/meshtastic-poi/pkg/engine"
)

var mergeCmd = &cobra.Command{
	Use:   "merge [files...]",
	Short: "Merge multiple POI datasets into one",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		setupLogging(verbose)
		var slices [][]*model.POI
		for _, path := range args {
			pois, err := loadPOIs(path)
			if err != nil {
				log.Fatal().Err(err).Str("file", path).Msg("read input")
			}
			slices = append(slices, pois)
		}

		merged := engine.Merge(slices...)
		out := outputPath
		if out == "" {
			out = "-"
		}
		if err := writePOIs(out, merged); err != nil {
			log.Fatal().Err(err).Msg("write output")
		}
		log.Info().Int("pois", len(merged)).Msg("merge complete")
	},
}
