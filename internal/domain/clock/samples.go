package clock

import (
	"time"

	"github.com/oernster/ribbonkit/domain/localtime"
)

// weekdayCycleYears is how long the calendar takes to bring every date back to the same weekday
// between 1901 and 2099, where every fourth year is a leap year: within it every day of every month
// falls on every weekday, so no pairing of weekday, day and month is missing from the samples.
const weekdayCycleYears = 28

// Samples answers every distinct time a cell can show in format (the kit's) with every distinct date
// it can show in dateFormat, the dates over the weekday cycle beginning on the first of from's year.
// They are what the page measures, in the font it really draws with, to find how wide a cell must be
// for its widest time and date (FR-620).
func Samples(from time.Time, format localtime.Format, dateFormat DateFormat) (times, dates []string) {
	start := time.Date(from.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
	writeDate := func(at time.Time) string { return at.Format(dateLayouts[dateFormat]) }
	dates = localtime.Distinct(start, start.AddDate(weekdayCycleYears, 0, 0), func(at time.Time) time.Time { return at.AddDate(0, 0, 1) }, writeDate)
	return localtime.TimeSamples(format), dates
}
