package zones

import (
	"testing"

	"github.com/oernster/timeribbon/internal/domain/sun"
)

func newZones(t *testing.T) *Zones {
	t.Helper()
	zones, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return zones
}

// CON-5, FR-302: the catalogue was generated from an older tz release than Go embeds, so every
// place it offers must still resolve here.
func TestEveryPlaceResolvesInTheEmbeddedDatabase(t *testing.T) {
	t.Parallel()
	zones := newZones(t)
	catalogue := zones.Catalogue()
	if len(catalogue) < 400 {
		t.Fatalf("only %d places; the catalogue did not load", len(catalogue))
	}
	for _, place := range catalogue {
		if _, err := zones.Resolve(place.Zone); err != nil {
			t.Errorf("%s: %v", place.Zone, err)
		}
		if place.Label == "" || place.Country == "" {
			t.Errorf("%s has an empty label or country: %+v", place.Zone, place)
		}
	}
}

func TestTheSearchExamplesAreInTheCatalogue(t *testing.T) {
	t.Parallel()
	want := map[string]string{"America/New_York": "New York", "Asia/Kolkata": "Kolkata", "Europe/Oslo": "Oslo"}
	for _, place := range newZones(t).Catalogue() {
		if label, ok := want[place.Zone]; ok && label == place.Label {
			delete(want, place.Zone)
		}
	}
	if len(want) != 0 {
		t.Errorf("missing %v", want)
	}
}

// FR-705: an id the database does not know is refused; so are the two Go would quietly accept.
func TestAnUnknownZoneIsRefused(t *testing.T) {
	t.Parallel()
	zones := newZones(t)
	for _, zone := range []string{"Not/AZone", "", "Local", "local"} {
		if _, err := zones.Resolve(zone); err == nil {
			t.Errorf("%q resolved", zone)
		}
	}
}

func TestAResolvedZoneIsCached(t *testing.T) {
	t.Parallel()
	zones := newZones(t)
	first, err := zones.Resolve("Europe/London")
	if err != nil {
		t.Fatal(err)
	}
	second, _ := zones.Resolve("Europe/London")
	if first != second {
		t.Error("the second resolve loaded the zone again")
	}
}

func TestTheCatalogueIsACopy(t *testing.T) {
	t.Parallel()
	zones := newZones(t)
	zones.Catalogue()[0].Zone = "changed"
	if zones.Catalogue()[0].Zone == "changed" {
		t.Error("a caller changed the catalogue")
	}
}

func TestAMalformedCatalogueLineIsRefused(t *testing.T) {
	t.Parallel()
	if _, err := fromText("# header\nEurope/London\n"); err == nil {
		t.Error("a line with no country was accepted")
	}
	if _, err := fromText("# header\nEurope/London\tBritain (UK)\tnorth\t0\n"); err == nil {
		t.Error("a line with no coordinates was accepted")
	}
	got, err := parse("# header\r\nEurope/London\tBritain (UK)\t51.5083\t-0.1253\r\n\r\n")
	if err != nil || len(got) != 1 || got[0].Country != "Britain (UK)" || got[0].At != (sun.Point{Latitude: 51.5083, Longitude: -0.1253}) {
		t.Errorf("got %+v (%v)", got, err)
	}
}

// FR-908: every place in the catalogue carries its zone city's coordinate, on the Earth; London's is
// where the tz database puts it.
func TestEveryPlaceHasItsZonesCoordinate(t *testing.T) {
	t.Parallel()
	for _, place := range newZones(t).Catalogue() {
		if place.At.Latitude < -90 || place.At.Latitude > 90 || place.At.Longitude < -180 || place.At.Longitude > 180 || place.At == (sun.Point{}) {
			t.Errorf("%s at %+v", place.Zone, place.At)
		}
		if place.Zone == "Europe/London" && place.At != (sun.Point{Latitude: 51.5083, Longitude: -0.1253}) {
			t.Errorf("London at %+v", place.At)
		}
	}
}
