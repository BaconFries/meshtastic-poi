package cli

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/BaconFries/meshtastic-poi/internal/providers/register"
)

var providersCmd = &cobra.Command{
	Use:   "providers",
	Short: "List available data providers",
	Run: func(cmd *cobra.Command, args []string) {
		setupLogging(verbose)
		registry := register.DefaultRegistry()
		types := registry.Types()
		sort.Strings(types)
		fmt.Println("Available providers:")
		for _, t := range types {
			fmt.Printf("  - %s\n", t)
		}
	},
}
