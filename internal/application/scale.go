package application

import (
	"fmt"

	"github.com/oernster/timeribbon/internal/domain/settings"
)

// SetScale chooses how large the clocks are drawn on top of their size in percent, then keeps it;
// any preview ends (FR-623). A value outside settings.MinScale to settings.MaxScale is refused and
// changes nothing.
func (s *Service) SetScale(percent int) error {
	err := choose(s, percent, func(c *settings.Settings) *int { return &c.Scale })
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.previewScale = 0
	return err
}

// PreviewScale draws the ribbon at percent without keeping it, while its grip is dragged; SetScale
// keeps the scale the drag ends at (FR-623). A value outside the bounds is refused.
func (s *Service) PreviewScale(percent int) error {
	if percent < settings.MinScale || percent > settings.MaxScale {
		return fmt.Errorf("%w: a scale of %d percent", ErrUnknownChoice, percent)
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.previewScale = percent
	return nil
}

// scaleOf answers the scale the ribbon is drawn at under current: the preview while a drag is under
// way, else the one kept. The caller holds the mutex.
func (s *Service) scaleOf(current settings.Settings) int {
	if s.previewScale != 0 {
		return s.previewScale
	}
	return current.Scale
}
