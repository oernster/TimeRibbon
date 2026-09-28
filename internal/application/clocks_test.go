package application

import (
	"errors"
	"slices"
	"testing"

	"github.com/oernster/timeribbon/internal/domain/settings"
)

// FR-301.
func TestAddingAClockAppendsItWithTheDefaultLabel(t *testing.T) {
	t.Parallel()
	r := newRig(t, withEntries(settings.Entry{ID: "old", Zone: "Europe/London", Label: "London"}))
	id, err := r.service.AddClock("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	saved := r.store.last(t)
	if len(saved.Clocks) != 2 || saved.Clocks[1] != (settings.Entry{ID: id, Zone: "America/New_York", Label: "New York"}) {
		t.Errorf("saved %+v", saved.Clocks)
	}
}

func TestAnUnknownZoneIsRefusedAndNothingIsSaved(t *testing.T) {
	t.Parallel()
	r := newRig(t, withEntries(settings.Entry{ID: "a", Zone: "Europe/London", Label: "London"}))
	if _, err := r.service.AddClock("Not/AZone"); !errors.Is(err, ErrUnknownZone) {
		t.Errorf("add: got %v", err)
	}
	if err := r.service.RezoneClock("a", "Not/AZone"); !errors.Is(err, ErrUnknownZone) {
		t.Errorf("rezone: got %v", err)
	}
	if len(r.store.saved) != 0 {
		t.Errorf("saved %d times", len(r.store.saved))
	}
}

// FR-303, FR-307.
func TestRenamingTrimsAndFallsBackToTheDefault(t *testing.T) {
	t.Parallel()
	r := newRig(t, withEntries(settings.Entry{ID: "a", Zone: "Europe/London", Label: "London"}))
	if err := r.service.RenameClock("a", "  Brighton "); err != nil {
		t.Fatal(err)
	}
	if got := r.store.last(t).Clocks[0].Label; got != "Brighton" {
		t.Errorf("got %q", got)
	}
	if err := r.service.RenameClock("a", " "); err != nil {
		t.Fatal(err)
	}
	if got := r.store.last(t).Clocks[0].Label; got != "London" {
		t.Errorf("an empty label should fall back to London, got %q", got)
	}
}

// FR-304, FR-706.
func TestChangingZoneKeepsACustomLabelAndRepairsAnInvalidClock(t *testing.T) {
	t.Parallel()
	r := newRig(t, withEntries(
		settings.Entry{ID: "a", Zone: "Europe/London", Label: "Brighton"},
		settings.Entry{ID: "b", Zone: "Not/AZone", Label: ""},
	))
	if err := r.service.RezoneClock("a", "Asia/Kolkata"); err != nil {
		t.Fatal(err)
	}
	if err := r.service.RezoneClock("b", "America/New_York"); err != nil {
		t.Fatal(err)
	}
	clocks := r.store.last(t).Clocks
	if clocks[0].Label != "Brighton" || clocks[0].Zone != "Asia/Kolkata" {
		t.Errorf("first: %+v", clocks[0])
	}
	if clocks[1].Label != "New York" || clocks[1].Zone != "America/New_York" {
		t.Errorf("repaired: %+v", clocks[1])
	}
}

// FR-305.
func TestRemovingPersistsTheRest(t *testing.T) {
	t.Parallel()
	r := newRig(t, withEntries(
		settings.Entry{ID: "a", Zone: "Europe/London"},
		settings.Entry{ID: "b", Zone: "Asia/Kolkata"},
		settings.Entry{ID: "c", Zone: "America/New_York"},
	))
	if err := r.service.RemoveClock("a"); err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, entry := range r.store.last(t).Clocks {
		ids = append(ids, entry.ID)
	}
	if !slices.Equal(ids, []string{"b", "c"}) {
		t.Errorf("got %v", ids)
	}
}

func TestAnUnknownClockIdChangesNothing(t *testing.T) {
	t.Parallel()
	r := newRig(t, withEntries(settings.Entry{ID: "a", Zone: "Europe/London"}))
	for name, err := range map[string]error{
		"rename": r.service.RenameClock("z", "x"),
		"rezone": r.service.RezoneClock("z", "Asia/Kolkata"),
		"remove": r.service.RemoveClock("z"),
	} {
		if !errors.Is(err, settings.ErrNoSuchClock) {
			t.Errorf("%s: got %v", name, err)
		}
	}
	if len(r.store.saved) != 0 {
		t.Errorf("saved %d times", len(r.store.saved))
	}
}

// FR-302: the acceptance examples, plus country and order.
func TestPlaceSearchMatchesLabelZoneOrCountry(t *testing.T) {
	t.Parallel()
	r := newRig(t, settings.Defaults())
	zones := func(places []Place) []string {
		var out []string
		for _, place := range places {
			out = append(out, place.Zone)
		}
		return out
	}
	cases := map[string][]string{
		"york":           {"America/New_York"},
		"KOLKATA":        {"Asia/Kolkata"},
		"united states":  {"America/Indiana/Indianapolis", "America/New_York"},
		"indiana/":       {"America/Indiana/Indianapolis"},
		"nowhere at all": nil,
	}
	for query, want := range cases {
		if got := zones(r.service.SearchPlaces(query)); !slices.Equal(got, want) {
			t.Errorf("%q: got %v, want %v", query, got, want)
		}
	}
	if got := len(r.service.SearchPlaces("  ")); got != 4 {
		t.Errorf("an empty query answers every place: got %d", got)
	}
}
