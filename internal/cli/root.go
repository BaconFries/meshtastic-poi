package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	configPath string
	verbose    bool
	outputPath string
)

// rootCmd is the base command.
var rootCmd = &cobra.Command{
	Use:   "meshtastic-poi",
	Short: "Offline POI management engine",
	Long: `meshtastic-poi is a provider-agnostic POI management engine.

Data flows through a canonical POI model with pluggable providers,
composable processing pipelines, and format exporters.`,
}

// Execute runs the CLI.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "config file path (YAML or JSON)")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "enable verbose logging")
	rootCmd.PersistentFlags().StringVarP(&outputPath, "output", "o", "", "output file path")
	rootCmd.PersistentFlags().StringVar(&outputFormat, "format", "geojson", "output format: geojson, meshtastic, csv")

	initCatalogPack()

	rootCmd.AddCommand(downloadCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(optimizeCmd)
	rootCmd.AddCommand(filterCmd)
	rootCmd.AddCommand(splitCmd)
	rootCmd.AddCommand(mergeCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(providersCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(indexCmd)
	rootCmd.AddCommand(catalogCmd)
	rootCmd.AddCommand(packCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(benchmarkCmd)
}

func exitOnError(err error) {
	if err != nil {
		os.Exit(1)
	}
}
