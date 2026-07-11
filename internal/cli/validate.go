package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/BaconFries/meshtastic-poi/internal/pipeline"
)

var (
	validateInput string
	validateJSON  bool
)

var validateCmd = &cobra.Command{
	Use:   "validate [file]",
	Short: "Validate POI data",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		setupLogging(verbose)
		pois, err := loadPOIs(args[0])
		if err != nil {
			log.Fatal().Err(err).Msg("read file")
		}

		report := pipeline.ValidateReport(pois)
		fmt.Println(pipeline.FormatReport(report))

		if validateJSON || outputPath != "" {
			out, err := json.MarshalIndent(report, "", "  ")
			if err != nil {
				log.Fatal().Err(err).Msg("marshal report")
			}
			if outputPath != "" {
				if err := os.WriteFile(outputPath, out, 0o644); err != nil {
					log.Fatal().Err(err).Msg("write report")
				}
			} else {
				fmt.Println(string(out))
			}
		}

		if !report.Valid {
			os.Exit(1)
		}
	},
}
