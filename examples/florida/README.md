# Florida Meshtastic POI Layers

This example downloads statewide Florida data and exports **one GeoJSON file per feature type**, so you can import each as a separate map layer in the Meshtastic app.

## Layers included

| Layer file | Source | Description |
|------------|--------|-------------|
| `fl-park-pois.geojson` | FDEP ArcGIS | State park points of interest |
| `fl-park-campgrounds.geojson` | FDEP ArcGIS | State park campsites |
| `fl-park-entrances.geojson` | FDEP ArcGIS | State park entrances |
| `fl-park-boundaries.geojson` | FDEP ArcGIS | State park boundaries (green filled polygons) |
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
# From repo root — Miami metro by default (~25.45–26.15°N, 80.90–80.05°W)
make florida-export

# Or run the script directly
./examples/florida/export-layers.sh
```

Output goes to `output/florida-miami/*.geojson` by default.

### Miami vs statewide

The export script defaults to a **Miami metro bounding box** so map layers stay usable on Meshtastic:

| Setting | Command | Output |
|---------|---------|--------|
| Miami (default) | `./examples/florida/export-layers.sh` | `output/florida-miami/` |
| Statewide | `BBOX= ./examples/florida/export-layers.sh` | `output/florida/` |
| Custom region | `BBOX="-80.5,25.7,-80.1,25.9" ./examples/florida/export-layers.sh` | `output/florida-miami/` |

Edit `examples/florida/config-miami.yaml` or the `BBOX` default in `export-layers.sh` to change the Miami box.

### Statewide split by county

For a full-state export without overloading the map, split each layer into one file per county:

```bash
BBOX= SPLIT_BY=county ./examples/florida/export-layers.sh
```

Output: `output/florida/by-county/fl-boat-ramps/MIAMI-DADE.geojson`, etc. Import only the counties you need.

## Map styling (`maplayer` format)

The `maplayer` exporter adds [simplestyle-spec](https://github.com/mapbox/simplestyle-spec) properties so POIs are visually distinct:

| POI type | Color | Example sources |
|----------|-------|-----------------|
| Campgrounds / cabins | Green | `POI_CLASSIFICATION`, recreation sites |
| Boat ramps / launches | Blue | `RampName`, ramp types |
| Beaches / dunes | Gold | recreation `FRESHWATER_BEACHES`, etc. |
| Trails | Purple | hiking, nature trail |
| Restrooms / bathhouses | Gray | park POI classification |
| Picnic / pavilions | Orange | park POI classification |
| Park boundaries | Green fill + outline | polygon geometry preserved |

**Names** use the specific POI (e.g. `Bathhouse Area 1 (Bathhouse)`) instead of only the park name.

**Meshtastic on Apple (iOS/macOS)** draws imported points as **colored circles** — it does not render `marker-symbol` icons or on-map text labels today. Color and size differences still help at zoom; full names live in the `name` property. **Park boundaries** render as filled green polygons on Apple and Android (Google).

## Import into Meshtastic

The Meshtastic app map layer importer expects standard **GeoJSON** (`FeatureCollection`). Use the `maplayer` export format:

```json
{
  "type": "FeatureCollection",
  "features": [
    {
      "type": "Feature",
      "geometry": { "type": "Point", "coordinates": [-80.156, 25.677] },
      "properties": {
        "name": "Bathhouse Area 1 (Bathhouse)",
        "category": "Bathhouse",
        "marker-color": "#7f8c8d",
        "marker-size": "small",
        "marker-symbol": "toilets"
      }
    }
  ]
}
```

Import one `.geojson` file per layer in your Meshtastic app so you can toggle feature types independently. Dense layers (e.g. boat ramps with thousands of points) only show pins once you zoom in to a region.

**Apple (iOS/macOS):** The Meshtastic app decodes feature `id` as an integer only. String ids (including numeric strings like `"1"`) pass file import but fail when the map loads the layer, so nothing appears. The `maplayer` exporter omits non-numeric ids and writes numeric ids as JSON numbers.

## Export a single layer

```bash
meshtastic-poi download "FL Boat Ramps" \
  --config examples/florida/config-miami.yaml \
  -o /tmp/boat.geojson

meshtastic-poi filter /tmp/boat.geojson \
  --bbox -80.90,25.45,-80.05,26.15 \
  -o /tmp/boat-miami.geojson

meshtastic-poi optimize /tmp/boat-miami.geojson \
  --dedupe --format maplayer \
  -o output/florida-miami/fl-boat-ramps.geojson
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

- **Park boundaries** export as green filled polygons when the source layer is `FL Park Boundaries`; other layers use styled point markers.
- **OSM queries** cover all of Florida and may take 1–3 minutes each; Overpass can rate-limit if run too quickly.
- Run `meshtastic-poi stats output/florida/fl-boat-ramps.geojson` after export to inspect counts.
