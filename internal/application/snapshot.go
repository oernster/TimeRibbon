package application

import (
	"cmp"
	"slices"
	"time"

	"github.com/oernster/ribbonkit/domain/ribbon"
	"github.com/oernster/timeribbon/internal/domain/clock"
	"github.com/oernster/timeribbon/internal/domain/settings"
)

// Words an invalid clock is shown with (FR-705, FR-706). They say what is wrong rather than
// relying on colour (NFR-U-2).
const (
	unknownZonePrefix = "Unknown time zone: "
	unreadablePrefix  = "This clock could not be read: "
)

// Cell is what one cell of the ribbon shows.
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

// Snapshot is everything the ribbon draws at one instant.
type Snapshot struct {
	Cells       []Cell
	Style       settings.Style
	Size        settings.Size
	Colour      ribbon.Colour
	Format      clock.Format
	DateFormat  clock.DateFormat
	Orientation ribbon.Orientation
	Theme       ribbon.Theme
	AlwaysOnTop bool
	// Opacity is how opaque the window is drawn, in percent; MinOpacity the least it may be, so the
	// page's control offers no value the setting would refuse (FR-622).
	Opacity    int
	MinOpacity int
	// Scale is the percent the ribbon is drawn at on top of its size, a preview's while its grip is
	// dragged; MinScale and MaxScale bound it (FR-623).
	Scale              float64
	MinScale, MaxScale int
	Layout             Layout
	// Now is the instant the snapshot was taken at.
	Now time.Time
	// NextRefresh is the minute boundary to take the next snapshot at (FR-208).
	NextRefresh time.Time
	// Notices are problems for the user to read, oldest first.
	Notices []string
	// SunMap is what the sun map draws (FR-905 to FR-908).
	SunMap SunMap
}

// timedCell is a cell with its zone's offset from UTC at the snapshot's instant; shown is false for
// a cell that cannot be shown, which has no offset to order by.
type timedCell struct {
	cell          Cell
	offsetSeconds int
	shown         bool
}

// behindGreenwich answers whether a zone's clock is behind UTC, which going east from Greenwich
// reaches last.
func behindGreenwich(offsetSeconds int) bool { return offsetSeconds < 0 }

// eastFromGreenwich orders cells starting at Greenwich and going east round the world: London,
// then Berlin, Tokyo, Melbourne, with New York last. Every place level with or ahead of UTC comes
// before every place behind it, each group by ascending offset, so UTC-10 never comes before UTC+14
// although the two keep the same time of day (FR-102). A cell that cannot be shown goes after every one
// that can.
func eastFromGreenwich(a, b timedCell) int {
	if a.shown != b.shown {
		if a.shown {
			return -1
		}
		return 1
	}
	if behind := behindGreenwich(a.offsetSeconds); behind != behindGreenwich(b.offsetSeconds) {
		if behind {
			return 1
		}
		return -1
	}
	return cmp.Compare(a.offsetSeconds, b.offsetSeconds)
}

// Snapshot answers what the ribbon shows now, one cell per clock ordered east from Greenwich, the
// reference; clocks keeping the same time keep the order they were added in (FR-102, FR-201 to
// FR-206, FR-612). The order is worked out at each snapshot, since daylight saving moves it. One clock
// that cannot be shown leaves every other one working (FR-705).
func (s *Service) Snapshot() Snapshot {
	now := s.ports.Clock.Now()
	s.mutex.Lock()
	defer s.mutex.Unlock()
	current := s.current.Normalised()
	timed := make([]timedCell, 0, len(current.Clocks))
	for _, entry := range current.Clocks {
		timed = append(timed, s.cell(entry, now, current.Format, current.DateFormat))
	}
	slices.SortStableFunc(timed, eastFromGreenwich)
	cells := make([]Cell, 0, len(timed))
	for _, each := range timed {
		cells = append(cells, each.cell)
	}
	return Snapshot{
		Cells:       cells,
		Style:       current.Style,
		Size:        current.Size,
		Colour:      current.Colour,
		Format:      current.Format,
		DateFormat:  current.DateFormat,
		Orientation: current.Orientation,
		Theme:       current.Theme,
		AlwaysOnTop: current.AlwaysOnTop,
		Opacity:     current.Opacity,
		MinOpacity:  ribbon.MinOpacity,
		Scale:       s.DrawnScale(current.Scale),
		MinScale:    ribbon.MinScale,
		MaxScale:    ribbon.MaxScale,
		Layout:      s.layoutFor(current),
		Now:         now,
		NextRefresh: clock.NextRefresh(now),
		Notices:     s.notices(),
		SunMap:      s.sunMap(current, cells, now),
	}
}

// cell answers one clock's cell at now, with the offset it is ordered by.
func (s *Service) cell(entry settings.Entry, now time.Time, format clock.Format, dateFormat clock.DateFormat) timedCell {
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
	reading := clock.Read(now, location, format, dateFormat)
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
