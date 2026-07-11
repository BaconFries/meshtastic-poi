package catalog

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// DefaultPath returns the per-user catalog file path.
func DefaultPath() (string, error) {
	home, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "meshtastic-poi", "catalog.yaml"), nil
}

// Dataset describes a managed POI dataset entry.
type Dataset struct {
	ID            string    `yaml:"id" json:"id"`
	Name          string    `yaml:"name" json:"name"`
	Provider      string    `yaml:"provider" json:"provider"`
	Version       string    `yaml:"version,omitempty" json:"version,omitempty"`
	URL           string    `yaml:"url,omitempty" json:"url,omitempty"`
	Checksum      string    `yaml:"checksum,omitempty" json:"checksum,omitempty"`
	LastUpdated   time.Time `yaml:"last_updated,omitempty" json:"last_updated,omitempty"`
	Tags          []string  `yaml:"tags,omitempty" json:"tags,omitempty"`
	OutputFormats []string  `yaml:"output_formats,omitempty" json:"output_formats,omitempty"`
}

// Catalog stores dataset metadata.
type Catalog struct {
	Datasets []Dataset `yaml:"datasets" json:"datasets"`
}

// Load reads a catalog from disk.
func Load(path string) (*Catalog, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Catalog{Datasets: []Dataset{}}, nil
		}
		return nil, err
	}
	var c Catalog
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parse catalog: %w", err)
	}
	if c.Datasets == nil {
		c.Datasets = []Dataset{}
	}
	return &c, nil
}

// Save writes the catalog to disk.
func (c *Catalog) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// Add inserts or replaces a dataset by ID.
func (c *Catalog) Add(d Dataset) error {
	if d.ID == "" {
		return fmt.Errorf("dataset id is required")
	}
	for i, existing := range c.Datasets {
		if existing.ID == d.ID {
			c.Datasets[i] = d
			return nil
		}
	}
	c.Datasets = append(c.Datasets, d)
	return nil
}

// Remove deletes a dataset by ID.
func (c *Catalog) Remove(id string) error {
	for i, d := range c.Datasets {
		if d.ID == id {
			c.Datasets = append(c.Datasets[:i], c.Datasets[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("dataset %q not found", id)
}

// Get returns a dataset by ID.
func (c *Catalog) Get(id string) (Dataset, bool) {
	for _, d := range c.Datasets {
		if d.ID == id {
			return d, true
		}
	}
	return Dataset{}, false
}
