package clock

import "time"

// weekdayCycleYears is how long the calendar takes to bring every date back to the same weekday
// between 1901 and 2099, where every fourth year is a leap year: within it every day of every month
// falls on every weekday, so no pairing of weekday, day and month is missing from the samples.
const weekdayCycleYears = 28

// hoursPerDay is the hours in a day, every one of which a cell may show.
const hoursPerDay = 24

// Samples answers every distinct time a cell can show in format and every distinct date it can show
// in dateFormat, the dates over the weekday cycle beginning on the first of from's year. They are
// what the page measures, in the font it really draws with, to find how wide a cell must be for its
// widest time and date (FR-620).
func Samples(from time.Time, format Format, dateFormat DateFormat) (times, dates []string) {
	start := time.Date(from.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
	layout := timeLayout(format)
	times = distinct(start, start.Add(hoursPerDay*time.Hour), func(at time.Time) time.Time { return at.Add(time.Minute) }, layout)
	dates = distinct(start, start.AddDate(weekdayCycleYears, 0, 0), func(at time.Time) time.Time { return at.AddDate(0, 0, 1) }, dateLayouts[dateFormat])
	return times, dates
}

// distinct answers each instant from start up to end, stepped by next, written in layout, once
// each in the order first met.
func distinct(start, end time.Time, next func(time.Time) time.Time, layout string) []string {
	seen := map[string]bool{}
	written := []string{}
	for at := start; at.Before(end); at = next(at) {
		text := at.Format(layout)
		if !seen[text] {
			seen[text] = true
			written = append(written, text)
		}
	}
	return written
}
