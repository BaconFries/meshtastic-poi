#!/usr/bin/env bash
# Export each Florida source as a separate Meshtastic POI layer.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
BIN="${ROOT}/bin/meshtastic-poi"

# Miami metro bbox: minLon,minLat,maxLon,maxLat
# Set BBOX= to export statewide (uses config.yaml and output/florida).
BBOX="${BBOX:--80.90,25.45,-80.05,26.15}"
if [[ -n "$BBOX" ]]; then
  CONFIG="${CONFIG:-${ROOT}/examples/florida/config-miami.yaml}"
  OUT="${OUT:-${ROOT}/output/florida-miami}"
else
  CONFIG="${CONFIG:-${ROOT}/examples/florida/config.yaml}"
  OUT="${OUT:-${ROOT}/output/florida}"
fi
TMP="${OUT}/.tmp"

mkdir -p "$OUT" "$TMP"

if [[ ! -x "$BIN" ]]; then
  echo "Building meshtastic-poi..."
  (cd "$ROOT" && go build -o "$BIN" ./cmd/meshtastic-poi)
fi

# Source names must match config.yaml exactly.
LAYERS=(
  "FL Park POIs"
  "FL Park Campgrounds"
  "FL Park Entrances"
  "FL Park Boundaries"
  "FL Recreation State"
  "FL Recreation County"
  "FL Recreation Federal"
  "FL Boat Ramps"
)

slug() {
  echo "$1" | tr '[:upper:]' '[:lower:]' | tr ' ' '-' | tr -cd 'a-z0-9-'
}

# Split statewide exports by field (e.g. county) into per-region files.
# Example: BBOX= SPLIT_BY=county ./examples/florida/export-layers.sh
SPLIT_BY="${SPLIT_BY:-}"

echo "Exporting ${#LAYERS[@]} Florida layers to ${OUT}/"
if [[ -n "$BBOX" ]]; then
  echo "Spatial filter (bbox): ${BBOX}"
fi
if [[ -n "$SPLIT_BY" ]]; then
  echo "Split by: ${SPLIT_BY} -> ${OUT}/by-${SPLIT_BY}/<layer>/"
fi
echo

for name in "${LAYERS[@]}"; do
  file="$(slug "$name").geojson"
  geo="${TMP}/$(slug "$name").geojson"

  echo "==> ${name}"
  "$BIN" download "$name" --config "$CONFIG" -o "$geo" --verbose

  input="$geo"
  if [[ -n "$BBOX" ]]; then
    filtered="${TMP}/$(slug "$name")-filtered.geojson"
    "$BIN" filter "$geo" --bbox "$BBOX" -o "$filtered"
    input="$filtered"
  fi

  if [[ -n "$SPLIT_BY" ]]; then
    split_dir="${OUT}/by-${SPLIT_BY}/$(slug "$name")"
    mkdir -p "$split_dir"
    "$BIN" split "$input" \
      --by "$SPLIT_BY" \
      --format maplayer \
      --output-dir "$split_dir"
    count=$(python3 -c "import json,glob; print(sum(len(json.load(open(f))['features']) for f in glob.glob('${split_dir}/*.geojson')))")
    parts=$(find "$split_dir" -name '*.geojson' | wc -l | tr -d ' ')
    echo "    ${count} features in ${parts} files -> ${split_dir}/"
  else
    "$BIN" optimize "$input" \
      --dedupe --remove-empty \
      --format maplayer \
      -o "${OUT}/${file}"

    count=$(python3 -c "import json; print(len(json.load(open('${OUT}/${file}'))['features']))")
    echo "    ${count} features -> ${OUT}/${file}"
  fi
  echo
done

echo "Done. Import each GeoJSON file as a separate map layer in the Meshtastic app."
echo "Layers written to: ${OUT}/"
if [[ -n "$SPLIT_BY" ]]; then
  echo "County (or ${SPLIT_BY}) splits: ${OUT}/by-${SPLIT_BY}/"
fi
