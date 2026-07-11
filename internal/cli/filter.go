package cli

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/BaconFries/meshtastic-poi/internal/output"
	"github.com/BaconFries/meshtastic-poi/internal/spatial"
)

var (
	filterLat    float64
	filterLon    float64
	filterRadius float64
	filterBBox   string
	filterPark   string
	filterCounty string
	filterName   string
	filterType   string
)

var filterCmd = &cobra.Command{
	Use:   "filter [file]",
	Short: "Filter POI data by spatial and attribute criteria",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		setupLogging(verbose)
		fc, err := output.ReadGeoJSON(args[0])
		if err != nil {
			log.Fatal().Err(err).Msg("read input")
		}

		result := fc
		if filterLat != 0 || filterLon != 0 {
			if filterRadius <= 0 {
				log.Fatal().Msg("--radius is required for radius filter")
			}
			result = spatial.FilterRadius(result, filterLat, filterLon, filterRadius)
		}
		if filterBBox != "" {
			bbox, err := spatial.ParseBBox(filterBBox)
			if err != nil {
				log.Fatal().Err(err).Msg("parse bbox")
			}
			result = spatial.FilterBBox(result, bbox[0], bbox[1], bbox[2], bbox[3])
		}

		attrs := map[string]string{}
		if filterPark != "" {
			attrs["park"] = filterPark
		}
		if filterCounty != "" {
			attrs["county"] = filterCounty
		}
		if filterName != "" {
			attrs["name"] = filterName
		}
		if filterType != "" {
			attrs["type"] = filterType
		}
		if len(attrs) > 0 {
			result = spatial.FilterAttributes(result, attrs)
		}

		out := outputPath
		if out == "" {
			out = "-"
		}
		if err := writeOutput(out, result); err != nil {
			log.Fatal().Err(err).Msg("write output")
		}
		log.Info().Int("features", len(result.Features)).Msg("filter complete")
	},
}

func init() {
	filterCmd.Flags().Float64Var(&filterLat, "lat", 0, "center latitude for radius filter")
	filterCmd.Flags().Float64Var(&filterLon, "lon", 0, "center longitude for radius filter")
	filterCmd.Flags().Float64Var(&filterRadius, "radius", 0, "radius in meters")
	filterCmd.Flags().StringVar(&filterBBox, "bbox", "", "bounding box: minLon,minLat,maxLon,maxLat")
	filterCmd.Flags().StringVar(&filterPark, "park", "", "filter by park name")
	filterCmd.Flags().StringVar(&filterCounty, "county", "", "filter by county name")
	filterCmd.Flags().StringVar(&filterName, "name", "", "filter by name")
	filterCmd.Flags().StringVar(&filterType, "type", "", "filter by type")
}
