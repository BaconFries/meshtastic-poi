package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"

	"github.com/BaconFries/meshtastic-poi/internal/providers"
)

// Config holds application configuration.
type Config struct {
	CacheDir string                   `mapstructure:"cache_dir"`
	Sources  []providers.SourceConfig `mapstructure:"sources"`
	Output   OutputConfig             `mapstructure:"output"`
}

// OutputConfig holds default output settings.
type OutputConfig struct {
	Dir      string `mapstructure:"dir"`
	Format   string `mapstructure:"format"`
	Filename string `mapstructure:"filename"`
}

var validate = validator.New()

// Default returns configuration with sensible defaults.
func Default() Config {
	home, _ := os.UserHomeDir()
	cacheDir := filepath.Join(home, ".cache", "meshtastic-poi")
	return Config{
		CacheDir: cacheDir,
		Output: OutputConfig{
			Format: "geojson",
		},
	}
}

// Load reads configuration from file and environment.
func Load(path string) (Config, error) {
	cfg := Default()
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetEnvPrefix("MESHTASTIC_POI")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if path != "" {
		v.SetConfigFile(path)
		if err := v.ReadInConfig(); err != nil {
			return cfg, err
		}
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return cfg, err
	}

	cfg.CacheDir = expandHome(cfg.CacheDir)
	if err := validate.Struct(cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func expandHome(path string) string {
	if path == "" {
		return path
	}
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}
