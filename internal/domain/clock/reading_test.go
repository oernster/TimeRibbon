package clock

import (
	"testing"
	"time"
	_ "time/tzdata"

	"github.com/oernster/ribbonkit/domain/localtime"
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

// FR-612: each date format writes the zone's own local date, the year included where it is shown;
// Kiritimati is already on the next day and in the next year.
func TestEachDateFormatWritesTheLocalDate(t *testing.T) {
	t.Parallel()
	at := instant(t, "2026-12-31T12:00:00Z")
	want := map[DateFormat][2]string{
		DayMonth:     {"Thursday, 31 December", "Friday, 1 January"},
		MonthDay:     {"Thursday, December 31", "Friday, January 1"},
		DayMonthYear: {"Thu 31/12/2026", "Fri 01/01/2027"},
		MonthDayYear: {"Thu 12/31/2026", "Fri 01/01/2027"},
		YearMonthDay: {"Thu 2026/12/31", "Fri 2027/01/01"},
	}
	if len(want) != len(DateFormats) {
		t.Fatalf("%d formats are offered but %d are checked", len(DateFormats), len(want))
	}
	for _, format := range DateFormats {
		losAngeles := Read(at, zone(t, "America/Los_Angeles"), localtime.TwentyFourHour, format).Date
		kiritimati := Read(at, zone(t, "Pacific/Kiritimati"), localtime.TwentyFourHour, format).Date
		if losAngeles != want[format][0] || kiritimati != want[format][1] {
			t.Errorf("%s wrote %q and %q, want %q", format, losAngeles, kiritimati, want[format])
		}
	}
}

// FR-201.
func TestLocalTimeInDistantZones(t *testing.T) {
	t.Parallel()
	at := instant(t, "2026-09-27T01:37:00Z")
	for name, want := range map[string]string{"America/New_York": "21:37", "Australia/Sydney": "11:37"} {
		if got := Read(at, zone(t, name), localtime.TwentyFourHour, DayMonth).Time; got != want {
			t.Errorf("%s: got %q, want %q", name, got, want)
		}
	}
}

// FR-202.
func TestLocalDateCrossesMidnightByZone(t *testing.T) {
	t.Parallel()
	at := instant(t, "2026-09-27T20:37:00Z")
	newYork := Read(at, zone(t, "America/New_York"), localtime.TwentyFourHour, DayMonth)
	sydney := Read(at, zone(t, "Australia/Sydney"), localtime.TwentyFourHour, DayMonth)
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
		got := Read(instant(t, each.at), newYork, localtime.TwentyFourHour, DayMonth)
		if got.Time != each.time || got.ZoneMark != each.mark || got.OffsetSeconds != int(each.offset.Seconds()) {
			t.Errorf("%s: got %s %s %d, want %s %s %v", each.at, got.Time, got.ZoneMark, got.OffsetSeconds, each.time, each.mark, each.offset)
		}
	}
}

// FR-205.
func TestYearBoundaryDiffersByZone(t *testing.T) {
	t.Parallel()
	at := instant(t, "2026-12-31T12:00:00Z")
	if got := Read(at, zone(t, "Pacific/Kiritimati"), localtime.TwentyFourHour, DayMonth).Date; got != "Friday, 1 January" {
		t.Errorf("Kiritimati: got %q", got)
	}
	if got := Read(at, zone(t, "America/Los_Angeles"), localtime.TwentyFourHour, DayMonth).Date; got != "Thursday, 31 December" {
		t.Errorf("Los Angeles: got %q", got)
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
		if got := Read(at, zone(t, name), localtime.TwentyFourHour, DayMonth).ZoneMark; got != want {
			t.Errorf("%s: got %q, want %q", name, got, want)
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
		got := Read(instant(t, each.at), time.UTC, localtime.TwentyFourHour, DayMonth)
		if got.HourAngle != each.hour || got.MinuteAngle != each.minutes {
			t.Errorf("%s: got hour %v minute %v, want %v %v", each.at, got.HourAngle, got.MinuteAngle, each.hour, each.minutes)
		}
	}
}
