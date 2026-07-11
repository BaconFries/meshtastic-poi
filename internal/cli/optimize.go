package cli

import (
	"strings"

	"github.com/paulmach/orb/geojson"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/BaconFries/meshtastic-poi/internal/optimizer"
	"github.com/BaconFries/meshtastic-poi/internal/output"
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
	Short: "Optimize and clean a GeoJSON dataset",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		setupLogging(verbose)
		fc, err := output.ReadGeoJSON(args[0])
		if err != nil {
			log.Fatal().Err(err).Msg("read input")
		}

		opts := optimizer.Options{
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
		result := optimizer.Pipeline(fc, opts)

		out := outputPath
		if out == "" {
			out = "-"
		}
		if err := writeOutput(out, result); err != nil {
			log.Fatal().Err(err).Msg("write output")
		}
		log.Info().Int("features", len(result.Features)).Msg("optimization complete")
	},
}

func init() {
	optimizeCmd.Flags().BoolVar(&optMinimal, "minimal", false, "enable minimal output mode")
	optimizeCmd.Flags().BoolVar(&optRemoveEmpty, "remove-empty", false, "remove empty properties")
	optimizeCmd.Flags().BoolVar(&optDedupe, "dedupe", false, "remove duplicate features")
	optimizeCmd.Flags().BoolVar(&optSortDistance, "sort-distance", false, "sort by distance from reference point")
	optimizeCmd.Flags().BoolVar(&optCompressProperties, "compress-properties", false, "keep only essential properties")
	optimizeCmd.Flags().StringVar(&optDropFields, "drop-fields", "", "comma-separated fields to drop")
	optimizeCmd.Flags().StringVar(&optKeepFields, "keep-fields", "", "comma-separated fields to keep")
	optimizeCmd.Flags().Float64Var(&optSortLat, "sort-lat", 0, "reference latitude for distance sort")
	optimizeCmd.Flags().Float64Var(&optSortLon, "sort-lon", 0, "reference longitude for distance sort")
	optimizeCmd.Flags().StringVar(&outputFormat, "format", "geojson", "output format: geojson or meshtastic")
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

func writeOutput(path string, fc *geojson.FeatureCollection) error {
	switch strings.ToLower(outputFormat) {
	case "meshtastic":
		return output.WriteMeshtastic(path, fc)
	default:
		return output.WriteGeoJSON(path, fc)
	}
}
