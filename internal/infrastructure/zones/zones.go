// Package zones lists the places the search offers and resolves zone ids (CON-5, FR-302). Resolving
// is the kit's, with the tz database built into the binary; see ribbonkit's infrastructure/zones.
package zones

import (
	_ "embed"
	"fmt"
	"strconv"
	"strings"
	"time"

	kitzones "github.com/oernster/ribbonkit/infrastructure/zones"
	"github.com/oernster/timeribbon/internal/application"
	"github.com/oernster/timeribbon/internal/domain/clock"
	"github.com/oernster/timeribbon/internal/domain/sun"
)

// places is the catalogue tools/genplaces writes: "zone<TAB>countries<TAB>latitude<TAB>longitude"
// per line, the coordinate being the zone's own city.
//
//go:embed places.tsv
var places string

// Zones is the application's Zones port.
type Zones struct {
	catalogue []application.Place
	resolver  kitzones.Resolver
}

// New parses the embedded catalogue. A failure is a defect in the build rather than on the
// machine, since the catalogue is built into the binary; a test holds it.
func New() (*Zones, error) {
	return fromText(places)
}

// fromText answers the zones over the catalogue text.
func fromText(text string) (*Zones, error) {
	catalogue, err := parse(text)
	if err != nil {
		return nil, err
	}
	return &Zones{catalogue: catalogue}, nil
}

// Resolve answers the location for zone through the kit's resolver: cached once loaded, with an empty
// id and "Local" refused (FR-705).
func (z *Zones) Resolve(zone string) (*time.Location, error) { return z.resolver.Resolve(zone) }

// Catalogue answers every place, in the catalogue's order.
func (z *Zones) Catalogue() []application.Place {
	return append([]application.Place(nil), z.catalogue...)
}

// parse reads the catalogue, skipping comment lines.
func parse(text string) ([]application.Place, error) {
	var out []application.Place
	for number, line := range strings.Split(text, "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != catalogueColumns || fields[0] == "" {
			return nil, fmt.Errorf("places.tsv line %d is not zone<TAB>country<TAB>latitude<TAB>longitude: %q", number+1, line)
		}
		latitude, errLatitude := strconv.ParseFloat(fields[2], 64)
		longitude, errLongitude := strconv.ParseFloat(fields[3], 64)
		if errLatitude != nil || errLongitude != nil {
			return nil, fmt.Errorf("places.tsv line %d has no coordinates: %q", number+1, line)
		}
		out = append(out, application.Place{
			Zone: fields[0], Label: clock.DefaultLabel(fields[0]), Country: fields[1],
			At: sun.Point{Latitude: latitude, Longitude: longitude},
		})
	}
	return out, nil
}

// catalogueColumns is how many columns each catalogue line holds.
const catalogueColumns = 4
