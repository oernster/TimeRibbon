package application

import (
	"fmt"

	"github.com/oernster/timeribbon/internal/domain/placement"
)

// Collapsed answers the arrangement of the tab an unpinned ribbon arranged as full shrinks to
// (FR-614): the band of placement.Tab on the edge it stands flush against, its thickness in that
// display's pixels. Nothing is saved: the stored place stays the full ribbon's. A ribbon flush
// against no edge never collapses (FR-619), so asking for its tab is refused.
func (s *Service) Collapsed(full Arrangement) (Arrangement, error) {
	if full.Edge == "" {
		return Arrangement{}, ErrNotAgainstAnEdge
	}
	monitors, err := s.ports.Monitors.Monitors()
	if err != nil {
		return Arrangement{}, fmt.Errorf("reading the displays: %w", err)
	}
	if len(monitors) == 0 {
		return Arrangement{}, ErrNoMonitors
	}
	monitor := mostOverlapped(monitors, full.At)
	thickness := placement.PixelsOf(placement.TabThickness, s.perDIP(monitor))
	band := placement.Tab(full.At, full.Size, full.Edge, thickness)
	return Arrangement{
		At:   placement.Point{X: band.Left, Y: band.Top},
		Size: placement.Size{Width: band.Width(), Height: band.Height()},
		DPI:  monitor.DPI,
	}, nil
}
