package clock

import (
	"slices"
	"testing"
	"time"

	"github.com/oernster/ribbonkit/domain/localtime"
)

// sampledFrom is any instant: the samples depend only on its year.
var sampledFrom = time.Date(2026, time.September, 29, 15, 4, 0, 0, time.UTC)

// FR-620: every minute of the day once, in the chosen time format.
func TestSamplesHoldEveryTimeOnce(t *testing.T) {
	t.Parallel()
	for format, want := range map[localtime.Format][]string{
		localtime.TwentyFourHour: {"00:00", "09:05", "12:30", "23:59"},
		localtime.TwelveHour:     {"12:00 AM", "9:05 AM", "12:30 PM", "11:59 PM"},
	} {
		times, _ := Samples(sampledFrom, format, DayMonth)
		if len(times) != hoursPerDay*minutesPerHour {
			t.Errorf("%s: %d times", format, len(times))
		}
		for _, each := range want {
			if !slices.Contains(times, each) {
				t.Errorf("%s: %q missing", format, each)
			}
		}
	}
}

// FR-620: a date in words pairs every weekday with every real day of every month (the widest, a
// Wednesday in September, among them) with nothing else; a date in numbers is every day of the cycle.
func TestSamplesHoldEveryPairingOfWeekdayDayAndMonth(t *testing.T) {
	t.Parallel()
	const realMonthDays, weekdays, commonYearDays, leapYears = 366, 7, 365, weekdayCycleYears / 4
	_, words := Samples(sampledFrom, localtime.TwentyFourHour, DayMonth)
	if len(words) != realMonthDays*weekdays {
		t.Errorf("day-month: %d dates, want %d", len(words), realMonthDays*weekdays)
	}
	for _, each := range []string{"Wednesday, 30 September", "Monday, 29 February", "Sunday, 1 January"} {
		if !slices.Contains(words, each) {
			t.Errorf("day-month: %q missing", each)
		}
	}
	if slices.Contains(words, "Monday, 31 September") {
		t.Error("day-month: a day September does not have")
	}
	_, numbers := Samples(sampledFrom, localtime.TwentyFourHour, DayMonthYear)
	if want := weekdayCycleYears*commonYearDays + leapYears; len(numbers) != want || numbers[0] != "Thu 01/01/2026" {
		t.Errorf("dmy: %d dates from %q, want %d from Thu 01/01/2026", len(numbers), numbers[0], want)
	}
}
