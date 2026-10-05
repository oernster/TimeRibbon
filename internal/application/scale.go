package application

import (
	"fmt"

	"github.com/oernster/timeribbon/internal/domain/settings"
	"github.com/oernster/timeribbon/ribbonkit/domain/placement"
	"github.com/oernster/timeribbon/ribbonkit/domain/ribbon"
)

// SetScale chooses how large the clocks are drawn on top of their size in percent, then keeps it;
// any preview ends and the sun map is no longer held (FR-623). A value outside ribbon.MinScale to
// ribbon.MaxScale is refused and changes nothing.
func (s *Service) SetScale(percent int) error {
	err := choose(s, percent, func(c *settings.Settings) *int { return &c.Scale })
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.previewScale, s.held = 0, heldMap{}
	return err
}

// PreviewScale draws the ribbon at percent without keeping it, while its grip is dragged; SetScale
// keeps the scale the drag ends at (FR-623). The percent need not be whole, so the ribbon follows
// the pointer pixel by pixel. The first preview of a drag holds the sun map where it stands. A
// value outside the bounds is refused.
func (s *Service) PreviewScale(percent float64) error {
	if !(percent >= ribbon.MinScale && percent <= ribbon.MaxScale) {
		return fmt.Errorf("%w: a scale of %v percent", ErrUnknownChoice, percent)
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.previewScale == 0 && s.last.known && s.last.sunMap != (placement.Rect{}) {
		ribbon := placement.Rect{Left: s.last.at.X, Top: s.last.at.Y, Right: s.last.at.X + s.last.size.Width, Bottom: s.last.at.Y + s.last.size.Height}
		s.held = heldMap{known: true, ribbon: ribbon, sunMap: s.last.sunMap}
	}
	s.previewScale = percent
	return nil
}

// scaleOf answers the scale the ribbon is drawn at under current: the preview while a drag is under
// way, else the one kept. The caller holds the mutex.
func (s *Service) scaleOf(current settings.Settings) float64 {
	if s.previewScale != 0 {
		return s.previewScale
	}
	return float64(current.Scale)
}

// heldOf answers the sun map held while the grip is dragged, read under the mutex.
func (s *Service) heldOf() heldMap {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.held
}
