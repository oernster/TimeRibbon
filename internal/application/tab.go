package application

import (
	"fmt"

	"github.com/oernster/timeribbon/internal/domain/placement"
	"github.com/oernster/timeribbon/internal/domain/settings"
)

// Collapsed answers the arrangement of the tab an unpinned ribbon arranged as full shrinks to
// (FR-614): the band of placement.Tab on the work area the ribbon overlaps most, its thickness in
// that display's pixels. Nothing is saved: the stored place stays the full ribbon's.
func (s *Service) Collapsed(full Arrangement) (Arrangement, error) {
	monitors, err := s.ports.Monitors.Monitors()
	if err != nil {
		return Arrangement{}, fmt.Errorf("reading the displays: %w", err)
	}
	if len(monitors) == 0 {
		return Arrangement{}, ErrNoMonitors
	}
	monitor := mostOverlapped(monitors, full.At)
	current := s.Settings()
	thickness := placement.PixelsOf(placement.TabThickness, s.perDIP(monitor))
	band := placement.Tab(full.At, full.Size, monitor.Work, current.Orientation == settings.Vertical, homeOf(current), thickness)
	return Arrangement{
		At:   placement.Point{X: band.Left, Y: band.Top},
		Size: placement.Size{Width: band.Width(), Height: band.Height()},
		DPI:  monitor.DPI,
	}, nil
}
