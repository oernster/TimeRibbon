package settings

import (
	"errors"
	"slices"
	"testing"

	"github.com/oernster/ribbonkit/domain/localtime"
	"github.com/oernster/ribbonkit/domain/ribbon"
	"github.com/oernster/timeribbon/internal/domain/clock"
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

// FR-703, FR-505: a first run is the kit's ribbon choices, then digital, large, 24-hour and the
// day-month date with no clocks.
func TestDefaultsAreDigitalTwentyFourHourVerticalAndNotOnTop(t *testing.T) {
	t.Parallel()
	got := Defaults()
	if got.Choices != ribbon.Defaults() {
		t.Errorf("the ribbon's choices are %+v, want the kit's defaults", got.Choices)
	}
	if got.Style != Digital || got.Size != Large || got.Format != localtime.TwentyFourHour ||
		got.DateFormat != clock.DayMonth || got.SunMap || got.PullOut || len(got.Clocks) != 0 {
		t.Errorf("got %+v", got)
	}
}

// A choice holding a word it does not offer is its default, the ribbon's own choices among them
// (normalised by the kit); a known one is kept.
func TestUnknownChoicesAreNormalisedToDefaults(t *testing.T) {
	t.Parallel()
	odd := Settings{Style: "sundial", Size: "huge", Format: "36h", DateFormat: "stardate"}
	odd.Orientation, odd.Opacity = "diagonal", 0
	got := odd.Normalised()
	want := Defaults()
	if got.Style != want.Style || got.Size != want.Size || got.Format != want.Format || got.DateFormat != want.DateFormat {
		t.Errorf("got %+v", got)
	}
	if got.Orientation != want.Orientation || got.Opacity != ribbon.MinOpacity {
		t.Errorf("the ribbon's choices were not normalised: %+v", got.Choices)
	}
	known := Settings{Style: Analogue, Size: Small, Format: localtime.TwelveHour, DateFormat: clock.YearMonthDay}
	if kept := known.Normalised(); kept.Style != Analogue || kept.Size != Small || kept.Format != localtime.TwelveHour ||
		kept.DateFormat != clock.YearMonthDay {
		t.Errorf("known choices were changed: %+v", kept)
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
