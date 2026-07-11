package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/BaconFries/meshtastic-poi/internal/pipeline"
)

var statsCmd = &cobra.Command{
	Use:   "stats [file]",
	Short: "Show statistics for a POI dataset",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		setupLogging(verbose)
		stats, err := pipeline.StatsFromFile(args[0])
		if err != nil {
			log.Fatal().Err(err).Msg("compute stats")
		}

		fmt.Print(pipeline.FormatStats(stats))

		if outputPath != "" {
			out, err := json.MarshalIndent(stats, "", "  ")
			if err != nil {
				log.Fatal().Err(err).Msg("marshal stats")
			}
			if err := os.WriteFile(outputPath, out, 0o644); err != nil {
				log.Fatal().Err(err).Msg("write stats")
			}
		}
	},
}
