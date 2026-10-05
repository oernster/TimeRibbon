// Package clock turns an instant and a zone into what one cell of the ribbon shows: the local
// time, the local weekday and date, the zone mark and the angles of an analogue dial's hands.
//
// It takes the instant as an argument and never reads the wall clock (FR-207). The zone arrives
// already resolved, so this package holds no tz database and no offset of its own (FR-204). How a
// time is written and the zone mark are every place-showing ribbon's, so ribbonkit's localtime
// package holds them (FR-203, FR-206).
package clock

import (
	"time"

	"github.com/oernster/ribbonkit/domain/localtime"
)

// DateFormat is how a date is written (FR-612): the day and month in words either way round; else the
// short weekday with the whole date in numbers in one of three orders.
type DateFormat string

// The date formats. The string values are what the settings file holds.
const (
	DayMonth     DateFormat = "day-month"
	MonthDay     DateFormat = "month-day"
	DayMonthYear DateFormat = "dmy"
	MonthDayYear DateFormat = "mdy"
	YearMonthDay DateFormat = "ymd"
)

// DateFormats lists the date formats in the order they are offered.
var DateFormats = []DateFormat{DayMonth, MonthDay, DayMonthYear, MonthDayYear, YearMonthDay}

// dateLayouts is the Go reference layout of each date format, its one home.
var dateLayouts = map[DateFormat]string{
	DayMonth:     "Monday, 2 January",
	MonthDay:     "Monday, January 2",
	DayMonthYear: "Mon 02/01/2006",
	MonthDayYear: "Mon 01/02/2006",
	YearMonthDay: "Mon 2006/01/02",
}

// Dial geometry. A full turn is 360 degrees; the hour hand makes one turn in twelve hours and
// the minute hand one in sixty minutes.
const (
	degreesPerTurn = 360
	hoursPerDial   = 12
	minutesPerHour = 60
)

// Reading is everything one cell shows for one instant.
type Reading struct {
	// Time is the local time in the chosen format, such as "21:37" or "9:37 PM".
	Time string
	// Date is the local date in the chosen date format, such as "Sunday, 27 September" or
	// "Sun 27/09/2026".
	Date string
	// ZoneMark is the zone's abbreviation where it has one of letters; else its UTC offset.
	ZoneMark string
	// HourAngle is the hour hand's angle in degrees clockwise from twelve.
	HourAngle float64
	// MinuteAngle is the minute hand's angle in degrees clockwise from twelve.
	MinuteAngle float64
	// OffsetSeconds is the zone's offset from UTC at the instant, daylight saving included: what
	// the ribbon is ordered by, east from Greenwich (FR-102).
	OffsetSeconds int
}

// Read answers what a cell shows at instant in location, writing times in format and dates in
// dateFormat. The settings are normalised before they reach here, so dateFormat is always known.
func Read(instant time.Time, location *time.Location, format localtime.Format, dateFormat DateFormat) Reading {
	local := instant.In(location)
	abbreviation, offset := local.Zone()
	minutes := float64(local.Minute())
	return Reading{
		Time:          localtime.Text(local, format),
		Date:          local.Format(dateLayouts[dateFormat]),
		ZoneMark:      localtime.ZoneMark(abbreviation, offset),
		HourAngle:     float64(local.Hour()%hoursPerDial)*degreesPerTurn/hoursPerDial + minutes*degreesPerTurn/(hoursPerDial*minutesPerHour),
		MinuteAngle:   minutes * degreesPerTurn / minutesPerHour,
		OffsetSeconds: offset,
	}
}
