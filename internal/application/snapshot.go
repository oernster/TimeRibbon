package application

import (
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
	ID          string
	Label       string
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

// Snapshot answers what the strip shows now, one cell per clock in order (FR-102, FR-201 to
// FR-206). One clock that cannot be shown leaves every other one working (FR-705).
func (s *Service) Snapshot() Snapshot {
	now := s.ports.Clock.Now()
	s.mutex.Lock()
	defer s.mutex.Unlock()
	current := s.current
	cells := make([]Cell, 0, len(current.Clocks))
	for _, entry := range current.Clocks {
		cells = append(cells, s.cell(entry, now, current.Format))
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

// cell answers one clock's cell at now.
func (s *Service) cell(entry settings.Entry, now time.Time, format clock.Format) Cell {
	label := entry.Label
	if label == "" {
		label = entry.Zone
	}
	if entry.Unreadable != "" {
		return Cell{ID: entry.ID, Label: label, Problem: unreadablePrefix + entry.Unreadable}
	}
	location, err := s.ports.Zones.Resolve(entry.Zone)
	if err != nil {
		return Cell{ID: entry.ID, Label: label, Problem: unknownZonePrefix + entry.Zone}
	}
	reading := clock.Read(now, location, format)
	return Cell{
		ID:          entry.ID,
		Label:       label,
		ZoneMark:    reading.ZoneMark,
		Time:        reading.Time,
		Date:        reading.Date,
		HourAngle:   reading.HourAngle,
		MinuteAngle: reading.MinuteAngle,
	}
}
