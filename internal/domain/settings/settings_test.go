package settings

import (
	"errors"
	"slices"
	"testing"

	"github.com/oernster/timeribbon/internal/domain/clock"
	"github.com/oernster/timeribbon/ribbonkit/domain/placement"
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

// FR-703, FR-103, FR-505, FR-613.
func TestDefaultsAreDigitalTwentyFourHourVerticalAndNotOnTop(t *testing.T) {
	t.Parallel()
	got := Defaults()
	if got.Style != Digital || got.Size != Large || got.Colour != Classic || got.Format != clock.TwentyFourHour ||
		got.DateFormat != clock.DayMonth || got.Orientation != Vertical || got.Theme != System || got.AlwaysOnTop || got.Placement != nil || len(got.Clocks) != 0 {
		t.Errorf("got %+v", got)
	}
	if !got.Pinned || got.OnTop(true) || got.LastEdge != nil {
		t.Errorf("a first run is pinned %v, on top %v, edge %v; want pinned, not on top, no edge", got.Pinned, got.OnTop(true), got.LastEdge)
	}
}

// FR-622: a first run is wholly opaque; an opacity outside the bounds, a hand-edited one or none at
// all, is brought to the nearer bound; one inside them is kept.
func TestOpacityIsHeldWithinItsBounds(t *testing.T) {
	t.Parallel()
	if got := Defaults().Opacity; got != MaxOpacity {
		t.Errorf("a first run is %d percent opaque", got)
	}
	for stored, want := range map[int]int{0: MinOpacity, MinOpacity - 1: MinOpacity, MinOpacity: MinOpacity, 55: 55, MaxOpacity: MaxOpacity, MaxOpacity + 1: MaxOpacity} {
		if got := (Settings{Opacity: stored}).Normalised().Opacity; got != want {
			t.Errorf("%d normalised to %d, want %d", stored, got, want)
		}
	}
}

// FR-623: a first run draws each size as it is; a scale outside the bounds is brought to the nearer.
func TestScaleIsHeldWithinItsBounds(t *testing.T) {
	t.Parallel()
	if got := Defaults().Scale; got != WholeScale {
		t.Errorf("a first run is scaled %d percent", got)
	}
	for stored, want := range map[int]int{0: MinScale, MinScale - 1: MinScale, 150: 150, MaxScale + 1: MaxScale} {
		if got := (Settings{Scale: stored}).Normalised().Scale; got != want {
			t.Errorf("%d normalised to %d, want %d", stored, got, want)
		}
	}
}

// FR-617, FR-619: a ribbon unpinned in effect (unpinned and flush) stays on top whatever Always on
// top holds; for one pinned (or unpinned away from every edge) Always on top decides.
func TestAnUnpinnedRibbonIsAlwaysOnTop(t *testing.T) {
	t.Parallel()
	for _, each := range []struct{ alwaysOnTop, pinned, flush, want bool }{
		{false, true, true, false}, {true, true, true, true}, {false, false, true, true}, {true, false, true, true},
		{false, false, false, false}, {true, false, false, true},
	} {
		s := Defaults()
		s.AlwaysOnTop, s.Pinned = each.alwaysOnTop, each.pinned
		if got := s.OnTop(each.flush); got != each.want {
			t.Errorf("Always on top %v, pinned %v, flush %v: on top %v, want %v", each.alwaysOnTop, each.pinned, each.flush, got, each.want)
		}
	}
}

// FR-619: the pin in effect; the choice itself is never changed by it.
func TestFlushnessGivesThePinInEffect(t *testing.T) {
	t.Parallel()
	for _, each := range []struct{ pinned, flush, want bool }{
		{true, true, true}, {true, false, true}, {false, true, false}, {false, false, true},
	} {
		s := Defaults()
		s.Pinned = each.pinned
		if got := s.PinnedInEffect(each.flush); got != each.want || s.Pinned != each.pinned {
			t.Errorf("pinned %v, flush %v: in effect %v, want %v", each.pinned, each.flush, got, each.want)
		}
	}
}

// FR-411: a remembered edge naming no edge, as a hand edit might, is forgotten; a real one is kept.
func TestAnUnknownRememberedEdgeIsForgotten(t *testing.T) {
	t.Parallel()
	s := Defaults()
	s.LastEdge = &placement.Against{Device: `\\.\DISPLAY1`, Edge: "middle"}
	if got := s.Normalised(); got.LastEdge != nil {
		t.Errorf("kept %+v", got.LastEdge)
	}
	s.LastEdge = &placement.Against{Device: `\\.\DISPLAY1`, Edge: placement.Left}
	if got := s.Normalised(); got.LastEdge == nil || *got.LastEdge != *s.LastEdge {
		t.Errorf("lost %+v", s.LastEdge)
	}
}

// FR-409: a horizontal ribbon goes to the top edge, a vertical one to the right.
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
	odd := Settings{
		Style: "sundial", Size: "huge", Colour: "mauve", Format: "36h", DateFormat: "stardate", Orientation: "diagonal", Theme: "sepia",
		AlwaysOnTop: true,
	}
	got := odd.Normalised()
	want := Defaults()
	want.AlwaysOnTop = true
	if got.Style != want.Style || got.Size != want.Size || got.Colour != want.Colour || got.Format != want.Format ||
		got.DateFormat != want.DateFormat || got.Orientation != want.Orientation || got.Theme != want.Theme || !got.AlwaysOnTop {
		t.Errorf("got %+v", got)
	}
	known := Settings{
		Style: Analogue, Size: Small, Colour: Neon, Format: clock.TwelveHour, DateFormat: clock.YearMonthDay, Orientation: Vertical, Theme: Dark,
	}
	if kept := known.Normalised(); kept.Style != Analogue || kept.Size != Small || kept.Colour != Neon || kept.Format != clock.TwelveHour ||
		kept.DateFormat != clock.YearMonthDay || kept.Orientation != Vertical || kept.Theme != Dark {
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
