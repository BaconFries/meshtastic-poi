#!/usr/bin/env bash
# Export each Florida source as a separate Meshtastic POI layer.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
CONFIG="${ROOT}/examples/florida/config.yaml"
OUT="${ROOT}/output/florida"
BIN="${ROOT}/bin/meshtastic-poi"
TMP="${ROOT}/output/florida/.tmp"

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

echo "Exporting ${#LAYERS[@]} Florida layers to ${OUT}/"
echo

for name in "${LAYERS[@]}"; do
  file="$(slug "$name").geojson"
  geo="${TMP}/$(slug "$name").geojson"

  echo "==> ${name}"
  "$BIN" download "$name" --config "$CONFIG" -o "$geo" --verbose

  "$BIN" optimize "$geo" \
    --dedupe --remove-empty \
    --format maplayer \
    -o "${OUT}/${file}"

  count=$(python3 -c "import json; print(len(json.load(open('${OUT}/${file}'))['features']))")
  echo "    ${count} features -> ${OUT}/${file}"
  echo
done

echo "Done. Import each GeoJSON file as a separate map layer in the Meshtastic app."
echo "Layers written to: ${OUT}/"
