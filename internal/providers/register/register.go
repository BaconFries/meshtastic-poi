package register

import (
	"github.com/BaconFries/meshtastic-poi/internal/providers"
	"github.com/BaconFries/meshtastic-poi/internal/providers/arcgis"
	csvprovider "github.com/BaconFries/meshtastic-poi/internal/providers/csv"
	"github.com/BaconFries/meshtastic-poi/internal/providers/garmin"
	geojsonprovider "github.com/BaconFries/meshtastic-poi/internal/providers/geojson"
	"github.com/BaconFries/meshtastic-poi/internal/providers/gpx"
	"github.com/BaconFries/meshtastic-poi/internal/providers/kml"
	"github.com/BaconFries/meshtastic-poi/internal/providers/osm"
)

// DefaultRegistry returns a registry with all built-in providers.
func DefaultRegistry() *providers.Registry {
	r := providers.NewRegistry()
	r.Register("arcgis", arcgis.New)
	r.Register("geojson", geojsonprovider.New)
	r.Register("csv", csvprovider.New)
	r.Register("osm", osm.New)
	r.Register("gpx", gpx.New)
	r.Register("kml", kml.New)
	r.Register("garmin", garmin.New)
	return r
}
