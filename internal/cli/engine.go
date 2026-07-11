package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/BaconFries/meshtastic-poi/internal/catalog"
	"github.com/BaconFries/meshtastic-poi/internal/config"
	"github.com/BaconFries/meshtastic-poi/internal/exporters"
	"github.com/BaconFries/meshtastic-poi/internal/model"
	"github.com/BaconFries/meshtastic-poi/internal/pack"
	"github.com/BaconFries/meshtastic-poi/internal/pipeline"
	"github.com/BaconFries/meshtastic-poi/internal/providers/register"
	"github.com/BaconFries/meshtastic-poi/internal/spatial"
	"github.com/BaconFries/meshtastic-poi/pkg/engine"
)

func writePOIs(path string, pois []*model.POI) error {
	format := outputFormat
	if format == "" {
		format = "geojson"
	}
	return exporters.WriteFile(path, format, pois)
}

func loadPOIs(path string) ([]*model.POI, error) {
	return engine.LoadGeoJSON(path)
}

func runPipeline(ctx context.Context, pois []*model.POI, opts pipeline.Options) ([]*model.POI, error) {
	return pipeline.Run(ctx, pois, pipeline.Default(opts))
}

var exportCmd = &cobra.Command{
	Use:   "export [file]",
	Short: "Export POI data to a target format",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		setupLogging(verbose)
		pois, err := loadPOIs(args[0])
		if err != nil {
			log.Fatal().Err(err).Msg("read input")
		}
		out := outputPath
		if out == "" {
			out = "-"
		}
		if err := writePOIs(out, pois); err != nil {
			log.Fatal().Err(err).Msg("export failed")
		}
	},
}

var indexCmd = &cobra.Command{
	Use:   "index [file]",
	Short: "Build a spatial index summary for a POI dataset",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		setupLogging(verbose)
		pois, err := loadPOIs(args[0])
		if err != nil {
			log.Fatal().Err(err).Msg("read input")
		}
		idx := engine.Index(pois)
		dupes := idx.FindDuplicateCoordinates(6)
		b := spatial.POIBound(pois)
		fmt.Printf("POIs indexed: %d\n", len(pois))
		fmt.Printf("Duplicate coordinate groups: %d\n", len(dupes))
		fmt.Printf("Bounds: %.6f,%.6f to %.6f,%.6f\n", b.Min[0], b.Min[1], b.Max[0], b.Max[1])
	},
}

var catalogCmd = &cobra.Command{
	Use:   "catalog",
	Short: "Manage the POI dataset catalog",
}

var catalogListCmd = &cobra.Command{
	Use:   "list",
	Short: "List catalog datasets",
	Run: func(cmd *cobra.Command, args []string) {
		path, err := catalog.DefaultPath()
		exitOnError(err)
		c, err := catalog.Load(path)
		exitOnError(err)
		for _, d := range c.Datasets {
			fmt.Printf("%s\t%s\t%s\n", d.ID, d.Provider, d.Name)
		}
	},
}

var catalogInfoCmd = &cobra.Command{
	Use:   "info [id]",
	Short: "Show catalog dataset details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path, err := catalog.DefaultPath()
		exitOnError(err)
		c, err := catalog.Load(path)
		exitOnError(err)
		d, ok := c.Get(args[0])
		if !ok {
			log.Fatal().Str("id", args[0]).Msg("dataset not found")
		}
		fmt.Printf("ID: %s\nName: %s\nProvider: %s\nURL: %s\n", d.ID, d.Name, d.Provider, d.URL)
	},
}

var catalogAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a dataset to the catalog",
	Run: func(cmd *cobra.Command, args []string) {
		id, _ := cmd.Flags().GetString("id")
		name, _ := cmd.Flags().GetString("name")
		provider, _ := cmd.Flags().GetString("provider")
		url, _ := cmd.Flags().GetString("url")
		path, err := catalog.DefaultPath()
		exitOnError(err)
		c, err := catalog.Load(path)
		exitOnError(err)
		exitOnError(c.Add(catalog.Dataset{
			ID:          id,
			Name:        name,
			Provider:    provider,
			URL:         url,
			LastUpdated: time.Now().UTC(),
		}))
		exitOnError(c.Save(path))
		fmt.Printf("Added dataset %s\n", id)
	},
}

var catalogRemoveCmd = &cobra.Command{
	Use:   "remove [id]",
	Short: "Remove a dataset from the catalog",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path, err := catalog.DefaultPath()
		exitOnError(err)
		c, err := catalog.Load(path)
		exitOnError(err)
		exitOnError(c.Remove(args[0]))
		exitOnError(c.Save(path))
	},
}

var packCmd = &cobra.Command{
	Use:   "pack",
	Short: "Build and manage POI packs",
}

var packBuildCmd = &cobra.Command{
	Use:   "build [pack.yaml]",
	Short: "Build a POI pack from catalog sources",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		setupLogging(verbose)
		cfg, err := config.Load(configPath)
		exitOnError(err)
		def, err := pack.Load(args[0])
		exitOnError(err)
		catPath, err := catalog.DefaultPath()
		exitOnError(err)
		cat, err := catalog.Load(catPath)
		exitOnError(err)
		pois, err := pack.Build(context.Background(), def, cat, cfg, outputPath)
		exitOnError(err)
		log.Info().Int("pois", len(pois)).Msg("pack build complete")
	},
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run environment and connectivity diagnostics",
	Run: func(cmd *cobra.Command, args []string) {
		setupLogging(verbose)
		fmt.Println("meshtastic-poi doctor")
		if cfgPath := configPath; cfgPath != "" {
			if _, err := os.Stat(cfgPath); err != nil {
				fmt.Printf("[FAIL] config: %v\n", err)
			} else {
				fmt.Printf("[OK] config: %s\n", cfgPath)
			}
		}
		catPath, err := catalog.DefaultPath()
		if err != nil {
			fmt.Printf("[FAIL] catalog path: %v\n", err)
		} else if _, err := os.Stat(catPath); err != nil {
			fmt.Printf("[WARN] catalog: not found at %s\n", catPath)
		} else {
			fmt.Printf("[OK] catalog: %s\n", catPath)
		}
		reg := register.DefaultRegistry()
		fmt.Printf("[OK] providers: %d registered\n", len(reg.Types()))
		exp := exporters.DefaultRegistry()
		fmt.Printf("[OK] exporters: %d registered\n", len(exp.Types()))
	},
}

var benchmarkCmd = &cobra.Command{
	Use:   "benchmark [file]",
	Short: "Benchmark import, validation, optimization, and export",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		setupLogging(verbose)
		start := time.Now()
		pois, err := loadPOIs(args[0])
		if err != nil {
			log.Fatal().Err(err).Msg("read input")
		}
		importDur := time.Since(start)

		start = time.Now()
		report := pipeline.ValidateReport(pois)
		validateDur := time.Since(start)

		start = time.Now()
		processed, err := runPipeline(context.Background(), pois, pipeline.Options{Minimal: true, Dedupe: true})
		exitOnError(err)
		optimizeDur := time.Since(start)

		start = time.Now()
		_ = exporters.WriteGeoJSONFile(os.DevNull, processed)
		exportDur := time.Since(start)

		idx := engine.Index(pois)
		start = time.Now()
		_ = idx.FindDuplicateCoordinates(6)
		dupDur := time.Since(start)

		fmt.Printf("POIs: %d\n", len(pois))
		fmt.Printf("Import: %v\n", importDur)
		fmt.Printf("Validation: %v (valid=%v)\n", validateDur, report.Valid)
		fmt.Printf("Optimization: %v\n", optimizeDur)
		fmt.Printf("Export: %v\n", exportDur)
		fmt.Printf("Duplicate detection: %v\n", dupDur)
	},
}

func initCatalogPack() {
	catalogAddCmd.Flags().String("id", "", "dataset id")
	catalogAddCmd.Flags().String("name", "", "dataset name")
	catalogAddCmd.Flags().String("provider", "", "provider type")
	catalogAddCmd.Flags().String("url", "", "source url")
	_ = catalogAddCmd.MarkFlagRequired("id")
	_ = catalogAddCmd.MarkFlagRequired("name")
	_ = catalogAddCmd.MarkFlagRequired("provider")

	catalogCmd.AddCommand(catalogListCmd, catalogInfoCmd, catalogAddCmd, catalogRemoveCmd)
	packCmd.AddCommand(packBuildCmd)
}
