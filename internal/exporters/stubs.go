package exporters

import (
	"context"
	"fmt"
	"io"

	"github.com/BaconFries/meshtastic-poi/internal/model"
)

// GPX exports POIs as GPX waypoints (stub: minimal implementation).
type GPX struct{}

func (GPX) Name() string { return "gpx" }

func (GPX) Export(_ context.Context, pois []*model.POI, w io.Writer) error {
	_, err := io.WriteString(w, `<?xml version="1.0" encoding="UTF-8"?>`+"\n<gpx version=\"1.1\">\n")
	if err != nil {
		return err
	}
	for _, p := range pois {
		if p == nil || !p.Valid() {
			continue
		}
		_, err = fmt.Fprintf(w, "  <wpt lat=\"%.7f\" lon=\"%.7f\"><name>%s</name></wpt>\n",
			p.Location[1], p.Location[0], xmlEscape(p.Name))
		if err != nil {
			return err
		}
	}
	_, err = io.WriteString(w, "</gpx>\n")
	return err
}

// KML exports POIs as KML placemarks (stub: minimal implementation).
type KML struct{}

func (KML) Name() string { return "kml" }

func (KML) Export(_ context.Context, pois []*model.POI, w io.Writer) error {
	_, err := io.WriteString(w, `<?xml version="1.0" encoding="UTF-8"?>`+"\n<kml xmlns=\"http://www.opengis.net/kml/2.2\"><Document>\n")
	if err != nil {
		return err
	}
	for _, p := range pois {
		if p == nil || !p.Valid() {
			continue
		}
		_, err = fmt.Fprintf(w, "  <Placemark><name>%s</name><Point><coordinates>%.7f,%.7f,0</coordinates></Point></Placemark>\n",
			xmlEscape(p.Name), p.Location[0], p.Location[1])
		if err != nil {
			return err
		}
	}
	_, err = io.WriteString(w, "</Document></kml>\n")
	return err
}

// Garmin exports POIs as Garmin Custom POI CSV (stub: minimal implementation).
type Garmin struct{}

func (Garmin) Name() string { return "garmin" }

func (Garmin) Export(_ context.Context, pois []*model.POI, w io.Writer) error {
	if _, err := io.WriteString(w, "Name,Description,Latitude,Longitude,Category\n"); err != nil {
		return err
	}
	for _, p := range pois {
		if p == nil || !p.Valid() {
			continue
		}
		_, err := fmt.Fprintf(w, "%q,%q,%.7f,%.7f,%q\n",
			p.Name, p.Description, p.Location[1], p.Location[0], p.Category)
		if err != nil {
			return err
		}
	}
	return nil
}

func xmlEscape(s string) string {
	r := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '&':
			r = append(r, "&amp;"...)
		case '<':
			r = append(r, "&lt;"...)
		case '>':
			r = append(r, "&gt;"...)
		case '"':
			r = append(r, "&quot;"...)
		default:
			r = append(r, s[i])
		}
	}
	return string(r)
}
