package cli

import (
	"context"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/BaconFries/meshtastic-poi/internal/pipeline"
)

var (
	optMinimal            bool
	optRemoveEmpty        bool
	optDedupe             bool
	optSortDistance       bool
	optCompressProperties bool
	optDropFields         string
	optKeepFields         string
	optSortLat            float64
	optSortLon            float64
	outputFormat          string
)

var optimizeCmd = &cobra.Command{
	Use:   "optimize [file]",
	Short: "Optimize and clean a POI dataset",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		setupLogging(verbose)
		pois, err := loadPOIs(args[0])
		if err != nil {
			log.Fatal().Err(err).Msg("read input")
		}

		opts := pipeline.Options{
			Minimal:            optMinimal,
			RemoveEmpty:        optRemoveEmpty,
			Dedupe:             optDedupe,
			SortDistance:       optSortDistance,
			CompressProperties: optCompressProperties,
			DropFields:         splitCSV(optDropFields),
			KeepFields:         splitCSV(optKeepFields),
			SortLat:            optSortLat,
			SortLon:            optSortLon,
		}
		result, err := runPipeline(context.Background(), pois, opts)
		if err != nil {
			log.Fatal().Err(err).Msg("pipeline failed")
		}

		out := outputPath
		if out == "" {
			out = "-"
		}
		if err := writePOIs(out, result); err != nil {
			log.Fatal().Err(err).Msg("write output")
		}
		log.Info().Int("pois", len(result)).Msg("optimization complete")
	},
}

func init() {
	optimizeCmd.Flags().BoolVar(&optMinimal, "minimal", false, "enable minimal output mode")
	optimizeCmd.Flags().BoolVar(&optRemoveEmpty, "remove-empty", false, "remove empty properties")
	optimizeCmd.Flags().BoolVar(&optDedupe, "dedupe", false, "remove duplicate POIs")
	optimizeCmd.Flags().BoolVar(&optSortDistance, "sort-distance", false, "sort by distance from reference point")
	optimizeCmd.Flags().BoolVar(&optCompressProperties, "compress-properties", false, "keep only essential properties")
	optimizeCmd.Flags().StringVar(&optDropFields, "drop-fields", "", "comma-separated fields to drop")
	optimizeCmd.Flags().StringVar(&optKeepFields, "keep-fields", "", "comma-separated fields to keep")
	optimizeCmd.Flags().Float64Var(&optSortLat, "sort-lat", 0, "reference latitude for distance sort")
	optimizeCmd.Flags().Float64Var(&optSortLon, "sort-lon", 0, "reference longitude for distance sort")
	optimizeCmd.Flags().StringVar(&outputFormat, "format", "geojson", "output format")
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
