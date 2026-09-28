// Package zones resolves zone ids through Go's time.LoadLocation and lists the places the search
// offers (CON-5, FR-302). The tz database is built into the binary through time/tzdata. Windows has
// no zone files of its own, so there the built-in rules are the only ones read; on macOS and Linux
// LoadLocation reads the system's zone files first and falls back to the built-in rules only when a
// zone is missing there.
package zones

import (
	_ "embed"
	"fmt"
	"strings"
	"sync"
	"time"
	_ "time/tzdata"

	"github.com/oernster/timeribbon/internal/application"
	"github.com/oernster/timeribbon/internal/domain/clock"
)

// places is the catalogue tools/genplaces writes: "zone<TAB>countries" per line.
//
//go:embed places.tsv
var places string

// Zones is the application's Zones port.
type Zones struct {
	catalogue []application.Place
	resolved  sync.Map
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

// Resolve answers the location for zone as the package comment describes, caching each zone once
// loaded. An empty id is refused rather than read as UTC, which is what Go would make of it.
func (z *Zones) Resolve(zone string) (*time.Location, error) {
	if cached, ok := z.resolved.Load(zone); ok {
		return cached.(*time.Location), nil
	}
	if zone == "" || strings.EqualFold(zone, "Local") {
		return nil, fmt.Errorf("%q is not a time zone id", zone)
	}
	location, err := time.LoadLocation(zone)
	if err != nil {
		return nil, err
	}
	z.resolved.Store(zone, location)
	return location, nil
}

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
		zone, country, ok := strings.Cut(line, "\t")
		if !ok || zone == "" {
			return nil, fmt.Errorf("places.tsv line %d is not zone<TAB>country: %q", number+1, line)
		}
		out = append(out, application.Place{Zone: zone, Label: clock.DefaultLabel(zone), Country: country})
	}
	return out, nil
}
