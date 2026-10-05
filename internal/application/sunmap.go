package application

import (
	"time"

	"github.com/oernster/ribbonkit/application/menus"
	"github.com/oernster/timeribbon/internal/domain/settings"
	"github.com/oernster/timeribbon/internal/domain/sun"
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
func (s *Service) sunMapItem() menus.Item {
	return menus.Item{Action: ActionSunMap, Label: labelSunMap, Checkable: true, Checked: s.Settings().SunMap}
}
