# Florida Meshtastic POI Layers

This example downloads statewide Florida data and exports **one GeoJSON file per feature type**, so you can import each as a separate map layer in the Meshtastic app.

## Layers included

| Layer file | Source | Description |
|------------|--------|-------------|
| `fl-park-pois.geojson` | FDEP ArcGIS | State park points of interest |
| `fl-park-campgrounds.geojson` | FDEP ArcGIS | State park campsites |
| `fl-park-entrances.geojson` | FDEP ArcGIS | State park entrances |
| `fl-park-boundaries.geojson` | FDEP ArcGIS | State park boundaries (centroid POIs) |
| `fl-recreation-state.geojson` | FDEP FORI | State-managed recreation sites |
| `fl-recreation-county.geojson` | FDEP FORI | County/municipal recreation sites |
| `fl-recreation-federal.geojson` | FDEP FORI | Federal recreation sites |
| `fl-boat-ramps.geojson` | FWC ArcGIS | Public boat ramps |
| `fl-osm-hospitals.geojson` | OSM Overpass | Hospitals statewide |
| `fl-osm-police.geojson` | OSM Overpass | Police stations |
| `fl-osm-fuel.geojson` | OSM Overpass | Fuel stations |
| `fl-osm-camp-sites.geojson` | OSM Overpass | Camp sites |
| `fl-osm-viewpoints.geojson` | OSM Overpass | Scenic viewpoints |
| `fl-osm-ranger-stations.geojson` | OSM Overpass | Ranger stations |
| `fl-osm-fire-stations.geojson` | OSM Overpass | Fire stations |

## Quick start

```bash
# From repo root
make florida-export

# Or run the script directly
./examples/florida/export-layers.sh
```

Output goes to `output/florida/*.geojson`.

## Import into Meshtastic

The Meshtastic app map layer importer expects standard **GeoJSON** (`FeatureCollection`) with **Point** features. Use the `maplayer` export format so each POI becomes a styled pin (polygons and building footprints are converted to their centroid):

```json
{
  "type": "FeatureCollection",
  "features": [
    {
      "type": "Feature",
      "geometry": { "type": "Point", "coordinates": [-81.3792, 28.5383] },
      "properties": {
        "name": "Park Name",
        "marker-color": "#c0392b",
        "marker-size": "medium"
      }
    }
  ]
}
```

Import one `.geojson` file per layer in your Meshtastic app so you can toggle feature types independently. Dense layers (e.g. boat ramps with thousands of points) only show pins once you zoom in to a region.

## Export a single layer

```bash
meshtastic-poi download "FL Boat Ramps" \
  --config examples/florida/config.yaml \
  -o /tmp/boat.geojson

meshtastic-poi optimize /tmp/boat.geojson \
  --dedupe --format maplayer \
  -o output/florida/fl-boat-ramps.geojson
```

## Add more layers

Edit `config.yaml` to add sources, then add the matching name to `LAYERS` in `export-layers.sh`.

Useful OSM tags for Florida backcountry/marine use:

- `amenity=shelter`
- `natural=peak`
- `leisure=marina`
- `emergency=assembly_point`

ArcGIS layers from [FDEP Open Data](https://ca.dep.state.fl.us/arcgis/rest/services/OpenData) and [FWC GIS](https://gis.myfwc.com/mapping/rest/services) can be added the same way.

## Notes

- **Polygon layers** (e.g. park boundaries) export as centroid POIs — good for map pins, not for area outlines.
- **OSM queries** cover all of Florida and may take 1–3 minutes each; Overpass can rate-limit if run too quickly.
- Run `meshtastic-poi stats output/florida/fl-boat-ramps.geojson` after export to inspect counts.
