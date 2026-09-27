// Package clock turns an instant and a zone into what one cell of the strip shows: the local
// time, the local weekday and date, the zone mark and the angles of an analogue dial's hands.
//
// It takes the instant as an argument and never reads the wall clock (FR-207). The zone arrives
// already resolved, so this package holds no tz database and no offset of its own (FR-204).
package clock

import (
	"fmt"
	"time"
	"unicode"
	"unicode/utf8"
)

// Format is how a time is written: 24-hour or 12-hour (FR-206).
type Format string

// The two formats. The string values are what the settings file holds.
const (
	TwentyFourHour Format = "24h"
	TwelveHour     Format = "12h"
)

// Go reference layouts for each piece of text a cell shows.
const (
	layoutTwentyFour = "15:04"
	layoutTwelve     = "3:04 PM"
	layoutDate       = "Monday, 2 January"
)

// Dial geometry. A full turn is 360 degrees; the hour hand makes one turn in twelve hours and
// the minute hand one in sixty minutes.
const (
	degreesPerTurn = 360
	hoursPerDial   = 12
	minutesPerHour = 60
	secondsPerHour = 3600
)

// offsetPrefix begins a zone mark written as an offset rather than an abbreviation.
const offsetPrefix = "UTC"

// Reading is everything one cell shows for one instant.
type Reading struct {
	// Time is the local time in the chosen format, such as "21:37" or "9:37 PM".
	Time string
	// Date is the local weekday, day and month, such as "Sunday, 27 September".
	Date string
	// ZoneMark is the zone's abbreviation where it has one of letters; else its UTC offset.
	ZoneMark string
	// HourAngle is the hour hand's angle in degrees clockwise from twelve.
	HourAngle float64
	// MinuteAngle is the minute hand's angle in degrees clockwise from twelve.
	MinuteAngle float64
}

// Read answers what a cell shows at instant in location, writing times in format.
func Read(instant time.Time, location *time.Location, format Format) Reading {
	local := instant.In(location)
	abbreviation, offset := local.Zone()
	layout := layoutTwentyFour
	if format == TwelveHour {
		layout = layoutTwelve
	}
	minutes := float64(local.Minute())
	return Reading{
		Time:        local.Format(layout),
		Date:        local.Format(layoutDate),
		ZoneMark:    ZoneMark(abbreviation, offset),
		HourAngle:   float64(local.Hour()%hoursPerDial)*degreesPerTurn/hoursPerDial + minutes*degreesPerTurn/(hoursPerDial*minutesPerHour),
		MinuteAngle: minutes * degreesPerTurn / minutesPerHour,
	}
}

// ZoneMark answers the text naming a zone beside its label (FR-203): the abbreviation the tz
// database gives where it begins with a letter; otherwise "UTC" with the signed offset in hours,
// minutes added only when there are some. An offset of zero is "UTC" alone.
func ZoneMark(abbreviation string, offsetSeconds int) string {
	first, _ := utf8.DecodeRuneInString(abbreviation)
	if unicode.IsLetter(first) {
		return abbreviation
	}
	if offsetSeconds == 0 {
		return offsetPrefix
	}
	sign := "+"
	if offsetSeconds < 0 {
		sign = "-"
		offsetSeconds = -offsetSeconds
	}
	hours := offsetSeconds / secondsPerHour
	minutes := offsetSeconds % secondsPerHour / minutesPerHour
	if minutes == 0 {
		return fmt.Sprintf("%s%s%d", offsetPrefix, sign, hours)
	}
	return fmt.Sprintf("%s%s%d:%02d", offsetPrefix, sign, hours, minutes)
}

// NextRefresh answers the next minute boundary after instant, computed from the instant itself
// rather than from the last refresh, so refreshes cannot drift (FR-208).
func NextRefresh(instant time.Time) time.Time {
	return instant.Truncate(time.Minute).Add(time.Minute)
}
