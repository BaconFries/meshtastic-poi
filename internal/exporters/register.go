package exporters

// DefaultRegistry returns all built-in exporters.
func DefaultRegistry() *Registry {
	r := NewRegistry()
	r.Register("geojson", GeoJSON{})
	r.Register("meshtastic", Meshtastic{})
	r.Register("csv", CSV{})
	r.Register("gpx", GPX{})
	r.Register("kml", KML{})
	r.Register("garmin", Garmin{})
	return r
}
