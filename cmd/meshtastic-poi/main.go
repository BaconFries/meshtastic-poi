package main

import (
	"os"

	"github.com/BaconFries/meshtastic-poi/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
