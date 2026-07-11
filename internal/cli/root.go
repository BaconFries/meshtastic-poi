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
	Short: "Download, validate, optimize, and export POI data for Meshtastic",
	Long: `meshtastic-poi is a cross-platform GIS toolkit for managing Points of Interest.

It supports multiple data providers (ArcGIS, GeoJSON, CSV, and more), spatial
filtering, optimization pipelines, and export to GeoJSON and Meshtastic formats.`,
}

// Execute runs the CLI.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "config file path (YAML)")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "enable verbose logging")
	rootCmd.PersistentFlags().StringVarP(&outputPath, "output", "o", "", "output file path")

	rootCmd.AddCommand(downloadCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(optimizeCmd)
	rootCmd.AddCommand(filterCmd)
	rootCmd.AddCommand(splitCmd)
	rootCmd.AddCommand(mergeCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(providersCmd)
}

func exitOnError(err error) {
	if err != nil {
		os.Exit(1)
	}
}
