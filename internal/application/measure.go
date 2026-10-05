package application

import (
	"fmt"

	"github.com/oernster/ribbonkit/domain/localtime"
	"github.com/oernster/ribbonkit/domain/placement"
	"github.com/oernster/timeribbon/internal/domain/clock"
	"github.com/oernster/timeribbon/internal/domain/settings"
)

// Measured is the width in DIP a cell needs to show its widest time and date whole, as the page
// measured it in the font it really draws with, together with the choices it was measured under:
// the size and style set the fonts, the two formats the words (FR-620).
type Measured struct {
	Size       settings.Size
	Style      settings.Style
	Format     localtime.Format
	DateFormat clock.DateFormat
	CellWidth  int
}

// TextSamples answers every time and date a cell can show under the current formats, for the page
// to measure (FR-620). Only the page can measure them: the font is whatever its web engine draws.
func (s *Service) TextSamples() (times, dates []string) {
	current := s.Settings()
	return clock.Samples(s.ports.Clock.Now(), current.Format, current.DateFormat)
}

// SetMeasured records the cell width the page measured the widest text to need; a width below
// zero is refused. Arranging the window afterwards gives the cells that width (FR-620).
func (s *Service) SetMeasured(measured Measured) error {
	if measured.CellWidth < 0 {
		return fmt.Errorf("%w: a cell of %d", placement.ErrNegativeLength, measured.CellWidth)
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.measured = measured
	return nil
}

// layoutFor answers the layout cells are drawn at under current: its size's own, with the style's
// clock cell widened to the width the page measured where that is wider and was measured under
// these very choices; a measurement for other choices waits for the page to measure again. The
// caller holds the mutex. Snapshot and arranging both read it, so the page's cells and the
// window's size cannot disagree (FR-610, FR-620).
func (s *Service) layoutFor(current settings.Settings) Layout {
	layout := s.layouts.For(current.Size)
	m := s.measured
	if m.Size != current.Size || m.Style != current.Style || m.Format != current.Format || m.DateFormat != current.DateFormat {
		return layout
	}
	if current.Style == settings.Analogue {
		layout.Analogue.Width = max(layout.Analogue.Width, m.CellWidth)
	} else {
		layout.Digital.Width = max(layout.Digital.Width, m.CellWidth)
	}
	return layout
}
