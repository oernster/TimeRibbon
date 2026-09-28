package clock

import (
	"testing"
	"time"
	_ "time/tzdata"
)

// zone loads a zone from the embedded tz database or fails the test.
func zone(t *testing.T, name string) *time.Location {
	t.Helper()
	location, err := time.LoadLocation(name)
	if err != nil {
		t.Fatalf("loading %s: %v", name, err)
	}
	return location
}

// instant parses an RFC 3339 instant or fails the test.
func instant(t *testing.T, text string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, text)
	if err != nil {
		t.Fatalf("parsing %s: %v", text, err)
	}
	return parsed
}

// FR-201.
func TestLocalTimeInDistantZones(t *testing.T) {
	t.Parallel()
	at := instant(t, "2026-09-27T01:37:00Z")
	for name, want := range map[string]string{"America/New_York": "21:37", "Australia/Sydney": "11:37"} {
		if got := Read(at, zone(t, name), TwentyFourHour).Time; got != want {
			t.Errorf("%s: got %q, want %q", name, got, want)
		}
	}
}

// FR-202.
func TestLocalDateCrossesMidnightByZone(t *testing.T) {
	t.Parallel()
	at := instant(t, "2026-09-27T20:37:00Z")
	newYork := Read(at, zone(t, "America/New_York"), TwentyFourHour)
	sydney := Read(at, zone(t, "Australia/Sydney"), TwentyFourHour)
	if newYork.Date != "Sunday, 27 September" || newYork.Time != "16:37" {
		t.Errorf("New York: got %q %q", newYork.Date, newYork.Time)
	}
	if sydney.Date != "Monday, 28 September" || sydney.Time != "06:37" {
		t.Errorf("Sydney: got %q %q", sydney.Date, sydney.Time)
	}
}

// FR-204.
func TestDaylightSavingTransitionIsFollowed(t *testing.T) {
	t.Parallel()
	newYork := zone(t, "America/New_York")
	// The offset follows the transition too, which is why the ribbon's order is worked out afresh
	// at each snapshot rather than stored (FR-102).
	cases := []struct {
		at, time, mark string
		offset         time.Duration
	}{
		{"2026-03-08T06:59:00Z", "01:59", "EST", -5 * time.Hour},
		{"2026-03-08T07:00:00Z", "03:00", "EDT", -4 * time.Hour},
	}
	for _, each := range cases {
		got := Read(instant(t, each.at), newYork, TwentyFourHour)
		if got.Time != each.time || got.ZoneMark != each.mark || got.OffsetSeconds != int(each.offset.Seconds()) {
			t.Errorf("%s: got %s %s %d, want %s %s %v", each.at, got.Time, got.ZoneMark, got.OffsetSeconds, each.time, each.mark, each.offset)
		}
	}
}

// FR-205.
func TestYearBoundaryDiffersByZone(t *testing.T) {
	t.Parallel()
	at := instant(t, "2026-12-31T12:00:00Z")
	if got := Read(at, zone(t, "Pacific/Kiritimati"), TwentyFourHour).Date; got != "Friday, 1 January" {
		t.Errorf("Kiritimati: got %q", got)
	}
	if got := Read(at, zone(t, "America/Los_Angeles"), TwentyFourHour).Date; got != "Thursday, 31 December" {
		t.Errorf("Los Angeles: got %q", got)
	}
}

// FR-206.
func TestTwelveAndTwentyFourHourFormats(t *testing.T) {
	t.Parallel()
	cases := []struct{ at, twentyFour, twelve string }{
		{"2026-09-27T06:37:00Z", "06:37", "6:37 AM"},
		{"2026-09-27T21:37:00Z", "21:37", "9:37 PM"},
		{"2026-09-27T00:00:00Z", "00:00", "12:00 AM"},
		{"2026-09-27T12:00:00Z", "12:00", "12:00 PM"},
	}
	for _, each := range cases {
		at := instant(t, each.at)
		if got := Read(at, time.UTC, TwentyFourHour).Time; got != each.twentyFour {
			t.Errorf("%s 24-hour: got %q, want %q", each.at, got, each.twentyFour)
		}
		if got := Read(at, time.UTC, TwelveHour).Time; got != each.twelve {
			t.Errorf("%s 12-hour: got %q, want %q", each.at, got, each.twelve)
		}
	}
}

// FR-203, with the values measured on 2026-09-27.
func TestZoneMarkPrefersLettersElseOffset(t *testing.T) {
	t.Parallel()
	at := instant(t, "2026-07-15T12:00:00Z")
	for name, want := range map[string]string{
		"America/New_York":  "EDT",
		"America/Sao_Paulo": "UTC-3",
		"Asia/Kathmandu":    "UTC+5:45",
		"Asia/Dubai":        "UTC+4",
	} {
		if got := Read(at, zone(t, name), TwentyFourHour).ZoneMark; got != want {
			t.Errorf("%s: got %q, want %q", name, got, want)
		}
	}
}

func TestZoneMarkForNumericForms(t *testing.T) {
	t.Parallel()
	cases := []struct {
		abbreviation string
		offset       int
		want         string
	}{
		{"-0930", -(9*secondsPerHour + 30*minutesPerHour), "UTC-9:30"},
		{"+00", 0, "UTC"},
		{"", 0, "UTC"},
		{"+1345", 13*secondsPerHour + 45*minutesPerHour, "UTC+13:45"},
	}
	for _, each := range cases {
		if got := ZoneMark(each.abbreviation, each.offset); got != each.want {
			t.Errorf("ZoneMark(%q, %d) = %q, want %q", each.abbreviation, each.offset, got, each.want)
		}
	}
}

// FR-603.
func TestHandAnglesForLocalTime(t *testing.T) {
	t.Parallel()
	cases := []struct {
		at            string
		hour, minutes float64
	}{
		{"2026-09-27T00:00:00Z", 0, 0},
		{"2026-09-27T03:00:00Z", 90, 0},
		{"2026-09-27T15:30:00Z", 105, 180},
		{"2026-09-27T21:37:00Z", 288.5, 222},
	}
	for _, each := range cases {
		got := Read(instant(t, each.at), time.UTC, TwentyFourHour)
		if got.HourAngle != each.hour || got.MinuteAngle != each.minutes {
			t.Errorf("%s: got hour %v minute %v, want %v %v", each.at, got.HourAngle, got.MinuteAngle, each.hour, each.minutes)
		}
	}
}

// FR-208.
func TestNextRefreshIsTheNextMinuteBoundary(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"2026-09-27T21:37:42.5Z": "2026-09-27T21:38:00Z",
		"2026-09-27T21:37:00Z":   "2026-09-27T21:38:00Z",
		"2026-12-31T23:59:59Z":   "2027-01-01T00:00:00Z",
	}
	for from, want := range cases {
		if got := NextRefresh(instant(t, from)); !got.Equal(instant(t, want)) {
			t.Errorf("after %s: got %s, want %s", from, got, want)
		}
	}
}
