package settings

import (
	"errors"
	"slices"
	"testing"

	"github.com/oernster/timeribbon/internal/domain/clock"
	"github.com/oernster/timeribbon/internal/domain/placement"
)

func withClocks(ids ...string) Settings {
	s := Defaults()
	for _, id := range ids {
		s = s.WithClockAdded(Entry{ID: id, Zone: "Europe/London", Label: "London"})
	}
	return s
}

func order(s Settings) []string {
	ids := make([]string, 0, len(s.Clocks))
	for _, entry := range s.Clocks {
		ids = append(ids, entry.ID)
	}
	return ids
}

// FR-703, FR-103, FR-505.
func TestDefaultsAreDigitalTwentyFourHourVerticalAndNotOnTop(t *testing.T) {
	t.Parallel()
	got := Defaults()
	if got.Style != Digital || got.Size != Large || got.Format != clock.TwentyFourHour || got.Orientation != Vertical ||
		got.Theme != System || got.AlwaysOnTop || got.Placement != nil || len(got.Clocks) != 0 {
		t.Errorf("got %+v", got)
	}
}

// FR-409: a horizontal strip goes to the top edge, a vertical one to the right.
func TestEachOrientationHasAHomeEdge(t *testing.T) {
	t.Parallel()
	for orientation, want := range map[Orientation]placement.Edge{Horizontal: placement.Top, Vertical: placement.Right} {
		if got, ok := HomeEdge(orientation); !ok || got != want {
			t.Errorf("%s: got %s, %v; want %s", orientation, got, ok, want)
		}
	}
	if _, ok := HomeEdge("diagonal"); ok {
		t.Error("an orientation the setting does not offer has a home edge")
	}
}

func TestUnknownChoicesAreNormalisedToDefaults(t *testing.T) {
	t.Parallel()
	odd := Settings{Style: "sundial", Size: "huge", Format: "36h", Orientation: "diagonal", Theme: "sepia", AlwaysOnTop: true}
	got := odd.Normalised()
	want := Defaults()
	want.AlwaysOnTop = true
	if got.Style != want.Style || got.Size != want.Size || got.Format != want.Format || got.Orientation != want.Orientation ||
		got.Theme != want.Theme || !got.AlwaysOnTop {
		t.Errorf("got %+v", got)
	}
	known := Settings{Style: Analogue, Size: Small, Format: clock.TwelveHour, Orientation: Vertical, Theme: Dark}
	if kept := known.Normalised(); kept.Style != Analogue || kept.Size != Small || kept.Format != clock.TwelveHour ||
		kept.Orientation != Vertical || kept.Theme != Dark {
		t.Errorf("known choices were changed: %+v", kept)
	}
	for _, theme := range []Theme{Light, System} {
		if got := (Settings{Theme: theme}).Normalised().Theme; got != theme {
			t.Errorf("theme %s was changed to %s", theme, got)
		}
	}
}

// FR-301.
func TestAddingAppendsAndLeavesTheOriginalAlone(t *testing.T) {
	t.Parallel()
	before := withClocks("a", "b")
	after := before.WithClockAdded(Entry{ID: "c"})
	if !slices.Equal(order(after), []string{"a", "b", "c"}) || len(before.Clocks) != 2 {
		t.Errorf("after %v, before %v", order(after), order(before))
	}
}

// FR-305.
func TestRemovingAClockClosesTheGap(t *testing.T) {
	t.Parallel()
	before := withClocks("a", "b", "c")
	after, err := before.WithoutClock("b")
	if err != nil || !slices.Equal(order(after), []string{"a", "c"}) || len(before.Clocks) != 3 {
		t.Errorf("after %v (%v), before %v", order(after), err, order(before))
	}
	if _, err := before.WithoutClock("z"); !errors.Is(err, ErrNoSuchClock) {
		t.Errorf("removing an unknown id: got %v", err)
	}
}

// FR-303.
func TestReplacingKeepsThePlaceInTheOrder(t *testing.T) {
	t.Parallel()
	start := withClocks("a", "b", "c")
	got, err := start.WithClockReplaced(Entry{ID: "b", Zone: "Asia/Tokyo", Label: "Tokyo"})
	if err != nil || got.Clocks[1].Label != "Tokyo" || start.Clocks[1].Label != "London" {
		t.Errorf("got %+v (%v); original %+v", got.Clocks[1], err, start.Clocks[1])
	}
	if _, err := start.WithClockReplaced(Entry{ID: "z"}); !errors.Is(err, ErrNoSuchClock) {
		t.Errorf("replacing an unknown id: got %v", err)
	}
}

func TestClockAnswersTheEntryWithThatId(t *testing.T) {
	t.Parallel()
	start := withClocks("a", "b")
	if got, err := start.Clock("b"); err != nil || got.ID != "b" {
		t.Errorf("got %+v (%v)", got, err)
	}
	if _, err := start.Clock("z"); !errors.Is(err, ErrNoSuchClock) {
		t.Errorf("an unknown id: got %v", err)
	}
}

// FR-304.
func TestChangingZoneKeepsACustomLabel(t *testing.T) {
	t.Parallel()
	custom := Rezoned(Entry{ID: "a", Zone: "Europe/London", Label: "Brighton"}, "Europe/Paris")
	if custom.Zone != "Europe/Paris" || custom.Label != "Brighton" {
		t.Errorf("a typed label was lost: %+v", custom)
	}
	followed := Rezoned(Entry{ID: "a", Zone: "Europe/London", Label: "London"}, "Europe/Paris")
	if followed.Label != "Paris" {
		t.Errorf("a default label did not follow the zone: %+v", followed)
	}
	repaired := Rezoned(Entry{ID: "a", Zone: "Not/AZone", Label: "", Unreadable: "bad", Original: "{}"}, "Asia/Tokyo")
	if repaired.Label != "Tokyo" || repaired.Unreadable != "" || repaired.Original != "" {
		t.Errorf("a repaired entry kept its fault: %+v", repaired)
	}
}
