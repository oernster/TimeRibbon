package application

import (
	"cmp"
	"slices"
	"time"

	"github.com/oernster/timestrip/internal/domain/clock"
	"github.com/oernster/timestrip/internal/domain/settings"
)

// Words an invalid clock is shown with (FR-705, FR-706). They say what is wrong rather than
// relying on colour (NFR-U-2).
const (
	unknownZonePrefix = "Unknown time zone: "
	unreadablePrefix  = "This clock could not be read: "
)

// Cell is what one cell of the strip shows.
type Cell struct {
	ID    string
	Label string
	// Zone is the zone id as stored, shown in Settings beside the label.
	Zone        string
	ZoneMark    string
	Time        string
	Date        string
	HourAngle   float64
	MinuteAngle float64
	// Problem is why the clock cannot be shown; empty for a working clock. A cell with a problem
	// shows its label and the problem, never another zone's time (FR-706).
	Problem string
}

// Snapshot is everything the strip draws at one instant.
type Snapshot struct {
	Cells       []Cell
	Style       settings.Style
	Format      clock.Format
	Orientation settings.Orientation
	Theme       settings.Theme
	AlwaysOnTop bool
	Layout      Layout
	// Now is the instant the snapshot was taken at.
	Now time.Time
	// NextRefresh is the minute boundary to take the next snapshot at (FR-208).
	NextRefresh time.Time
	// Notices are problems for the user to read, oldest first.
	Notices []string
}

// timedCell is a cell with its zone's offset from UTC at the snapshot's instant; shown is false for
// a cell that cannot be shown, which has no offset to order by.
type timedCell struct {
	cell          Cell
	offsetSeconds int
	shown         bool
}

// secondsPerDay is one turn of the world, which a place behind Greenwich is reached after going east.
var secondsPerDay = int((24 * time.Hour).Seconds())

// eastOfGreenwich answers how far east of Greenwich a zone's clock is, in seconds: its offset from
// UTC where that is ahead or level, else a whole day more, since going east from Greenwich reaches
// the places behind it last.
func eastOfGreenwich(offsetSeconds int) int {
	if offsetSeconds < 0 {
		return offsetSeconds + secondsPerDay
	}
	return offsetSeconds
}

// eastFromGreenwich orders cells starting at Greenwich and going east round the world: London,
// then Berlin, Tokyo, Melbourne, with New York last. A cell that cannot be shown goes after every
// one that can.
func eastFromGreenwich(a, b timedCell) int {
	if a.shown != b.shown {
		if a.shown {
			return -1
		}
		return 1
	}
	return cmp.Compare(eastOfGreenwich(a.offsetSeconds), eastOfGreenwich(b.offsetSeconds))
}

// Snapshot answers what the strip shows now, one cell per clock ordered east from Greenwich, the
// reference; clocks keeping the same time keep the order they were added in (FR-102, FR-201 to
// FR-206). The order is worked out at each snapshot, since daylight saving moves it. One clock
// that cannot be shown leaves every other one working (FR-705).
func (s *Service) Snapshot() Snapshot {
	now := s.ports.Clock.Now()
	s.mutex.Lock()
	defer s.mutex.Unlock()
	current := s.current
	timed := make([]timedCell, 0, len(current.Clocks))
	for _, entry := range current.Clocks {
		timed = append(timed, s.cell(entry, now, current.Format))
	}
	slices.SortStableFunc(timed, eastFromGreenwich)
	cells := make([]Cell, 0, len(timed))
	for _, each := range timed {
		cells = append(cells, each.cell)
	}
	return Snapshot{
		Cells:       cells,
		Style:       current.Style,
		Format:      current.Format,
		Orientation: current.Orientation,
		Theme:       current.Theme,
		AlwaysOnTop: current.AlwaysOnTop,
		Layout:      s.layout,
		Now:         now,
		NextRefresh: clock.NextRefresh(now),
		Notices:     s.notices(),
	}
}

// cell answers one clock's cell at now, with the offset it is ordered by.
func (s *Service) cell(entry settings.Entry, now time.Time, format clock.Format) timedCell {
	label := entry.Label
	if label == "" {
		label = entry.Zone
	}
	if entry.Unreadable != "" {
		return timedCell{cell: Cell{ID: entry.ID, Label: label, Zone: entry.Zone, Problem: unreadablePrefix + entry.Unreadable}}
	}
	location, err := s.ports.Zones.Resolve(entry.Zone)
	if err != nil {
		return timedCell{cell: Cell{ID: entry.ID, Label: label, Zone: entry.Zone, Problem: unknownZonePrefix + entry.Zone}}
	}
	reading := clock.Read(now, location, format)
	return timedCell{
		cell: Cell{
			ID:          entry.ID,
			Label:       label,
			Zone:        entry.Zone,
			ZoneMark:    reading.ZoneMark,
			Time:        reading.Time,
			Date:        reading.Date,
			HourAngle:   reading.HourAngle,
			MinuteAngle: reading.MinuteAngle,
		},
		offsetSeconds: reading.OffsetSeconds,
		shown:         true,
	}
}
