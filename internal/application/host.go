package application

import (
	"github.com/oernster/timeribbon/internal/domain/settings"
	"github.com/oernster/timeribbon/ribbonkit/application/arranger"
	"github.com/oernster/timeribbon/ribbonkit/domain/ribbon"
)

// host is the service as the arranger's host: the clocks are the ribbon's content and the ribbon's
// choices are saved with the rest of the settings, so a save that fails raises the same notice
// (FR-707). Not the service itself, so its two methods stay off the service's surface.
type host struct{ s *Service }

// Ribbon answers the ribbon's choices with its content, read together under one lock (FR-105): a
// cell for each notice, then for each clock, the prompt standing in for them when there are none
// (FR-107); each cell the size of the style's, widened to the measured text where that applies
// (FR-610, FR-620); the handle's lane added while the sun map is on (FR-903).
func (h host) Ribbon() (ribbon.Choices, arranger.Content) {
	s := h.s
	s.mutex.Lock()
	defer s.mutex.Unlock()
	current := s.current.Normalised()
	layout := s.layoutFor(current)
	cell := layout.Digital
	switch {
	case len(current.Clocks) == 0:
		cell = layout.Prompt
	case current.Style == settings.Analogue:
		cell = layout.Analogue
	}
	content := arranger.Content{
		Cell:    cell,
		Cells:   len(s.notices()) + max(len(current.Clocks), 1),
		Padding: layout.Padding,
		Map:     current.SunMap,
		PullOut: current.PullOut,
	}
	if current.SunMap {
		content.Lane = layout.HandleLane
	}
	return current.Choices, content
}

// ChangeRibbon applies edit to the ribbon's choices through the one save path, change.
func (h host) ChangeRibbon(edit func(ribbon.Choices) ribbon.Choices) error {
	return h.s.change(func(current settings.Settings) (settings.Settings, error) {
		current.Choices = edit(current.Choices)
		return current, nil
	})
}
