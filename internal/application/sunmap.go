package application

import (
	"time"

	"github.com/oernster/timeribbon/internal/domain/settings"
	"github.com/oernster/timeribbon/internal/domain/sun"
	"github.com/oernster/timeribbon/ribbonkit/domain/placement"
)

// SunMap is what the sun map draws at one instant (FR-905 to FR-908).
type SunMap struct {
	// On is whether the sun map is shown with the ribbon (FR-901); PullOut whether the map is pulled
	// out, whichever way the ribbon runs (FR-903).
	On, PullOut bool
	// Subsolar is where the sun stands overhead at the snapshot's instant.
	Subsolar sun.Point
	// Marks are the clocks' places, in the cells' order, one for each clock whose zone has one.
	Marks []Mark
}

// Mark is one clock's place on the sun map (FR-908).
type Mark struct {
	Label string
	At    sun.Point
}

// sunMap answers the sun map for cells at now: a mark for each cell whose zone's city the catalogue
// knows, none for a zone with no place there (such as UTC) or a clock that cannot be shown.
func (s *Service) sunMap(current settings.Settings, cells []Cell, now time.Time) SunMap {
	marks := []Mark{}
	if len(cells) == 0 {
		return SunMap{On: current.SunMap, PullOut: current.PullOut, Subsolar: sun.Subsolar(now), Marks: marks}
	}
	places := map[string]sun.Point{}
	for _, place := range s.ports.Zones.Catalogue() {
		places[place.Zone] = place.At
	}
	for _, cell := range cells {
		if at, ok := places[cell.Zone]; ok && cell.Problem == "" {
			marks = append(marks, Mark{Label: cell.Label, At: at})
		}
	}
	return SunMap{On: current.SunMap, PullOut: current.PullOut, Subsolar: sun.Subsolar(now), Marks: marks}
}

// SetSunMap turns the sun map on or off (FR-901). Arranging the window afterwards makes room for it.
func (s *Service) SetSunMap(on bool) error {
	return choose(s, on, func(c *settings.Settings) *bool { return &c.SunMap })
}

// SetPullOut opens or closes the pull out, one remembered choice for both orientations (FR-903).
func (s *Service) SetPullOut(open bool) error {
	return choose(s, open, func(c *settings.Settings) *bool { return &c.PullOut })
}

// sunMapItem follows Pin ribbon in both menus, ticked while the sun map is on (FR-901).
func (s *Service) sunMapItem() MenuItem {
	return MenuItem{Action: ActionSunMap, Label: labelSunMap, Checkable: true, Checked: s.Settings().SunMap}
}

// mapBeside answers where the sun map goes for a ribbon arranged at at of size on monitor, flush
// against edge (none when empty): the side it adjoins (where the pull out's handle goes too) with
// its rectangle; false for the rectangle while no map is shown, the pull out closed included
// (FR-902 to FR-904). The side is empty while the sun map is off.
func (s *Service) mapBeside(current settings.Settings, at placement.Point, size placement.Size, monitor placement.Monitor, edge placement.Edge) (placement.Edge, placement.Rect, bool) {
	if !current.SunMap {
		return "", placement.Rect{}, false
	}
	vertical := current.Orientation == settings.Vertical
	ribbon := placement.Rect{Left: at.X, Top: at.Y, Right: at.X + size.Width, Bottom: at.Y + size.Height}
	side := placement.InnerSide(ribbon, monitor.Work, vertical, edge)
	if !current.PullOut {
		return side, placement.Rect{}, false
	}
	perDIP := s.perDIP(monitor)
	minimum := placement.PixelsOf(placement.MapMinimumWidth, perDIP)
	floor := placement.PixelsOf(placement.MapFloor, perDIP)
	rect, shown := placement.MapBeside(ribbon, monitor.Work, side, minimum, floor)
	if held := s.heldOf(); held.known && shown {
		rect = placement.MapHeld(ribbon, side, held.sunMap, held.ribbon)
	}
	return side, rect, shown
}
