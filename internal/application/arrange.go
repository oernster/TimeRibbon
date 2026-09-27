package application

import (
	"fmt"
	"slices"

	"github.com/oernster/timestrip/internal/domain/placement"
	"github.com/oernster/timestrip/internal/domain/settings"
)

// Arrangement is where the window goes and how big it is, in physical pixels.
type Arrangement struct {
	At   placement.Point
	Size placement.Size
	// Scrolls is true when the cells need more room than the work area offers (FR-106).
	Scrolls bool
	// DPI is the DPI of the monitor the strip is on.
	DPI int
}

// Launch arranges the strip at launch: on its stored monitor and offset, else at the default place
// on the primary monitor, always wholly inside a work area (FR-403, FR-405).
func (s *Service) Launch() (Arrangement, error) {
	arranged, _, err := s.arrange(func(monitors []placement.Monitor, current settings.Settings) placement.Monitor {
		return storedOrPrimary(monitors, current.Placement)
	}, func(size placement.Size, monitors []placement.Monitor, current settings.Settings) placement.Placed {
		placed, _ := placement.Restore(current.Placement, monitors, size)
		return placed
	})
	return arranged, err
}

// Rearrange arranges a strip now at at after its content or the displays changed: the same
// top-left corner, moved the least distance that keeps it wholly inside the work area it overlaps
// most (FR-104, FR-406). Nothing is saved, so a monitor that comes back finds its placement kept.
func (s *Service) Rearrange(at placement.Point) (Arrangement, error) {
	arranged, _, err := s.recovered(at)
	return arranged, err
}

// Moved records where a drag left the strip (FR-404) and answers where it belongs, which differs
// only when the drag left part of it off every work area.
func (s *Service) Moved(at placement.Point) (Arrangement, error) {
	arranged, monitor, err := s.recovered(at)
	if err != nil {
		return arranged, err
	}
	stored := placement.Record(arranged.At, monitor)
	return arranged, s.change(func(current settings.Settings) (settings.Settings, error) {
		current.Placement = &stored
		return current, nil
	})
}

// Centred answers a window of size, in DIP, centred on the work area of the monitor holding at
// (CON-6): where Settings opens, since it shares the strip's window. Nothing is saved.
func (s *Service) Centred(at placement.Point, size placement.Size) (Arrangement, error) {
	monitors, err := s.ports.Monitors.Monitors()
	if err != nil {
		return Arrangement{}, fmt.Errorf("reading the displays: %w", err)
	}
	if len(monitors) == 0 {
		return Arrangement{}, ErrNoMonitors
	}
	monitor := mostOverlapped(monitors, at)
	work := monitor.Work
	// Never larger than the work area, so a short display still shows the whole surface.
	pixels := placement.Size{
		Width:  min(placement.Scale(size.Width, placement.BaseDPI, monitor.DPI), work.Width()),
		Height: min(placement.Scale(size.Height, placement.BaseDPI, monitor.DPI), work.Height()),
	}
	centre := placement.Point{X: work.Left + (work.Width()-pixels.Width)/2, Y: work.Top + (work.Height()-pixels.Height)/2}
	return Arrangement{At: placement.Clamp(centre, pixels, work), Size: pixels, DPI: monitor.DPI}, nil
}

// recovered answers the arrangement of a strip at at, clamped into the work area it overlaps most.
func (s *Service) recovered(at placement.Point) (Arrangement, placement.Monitor, error) {
	return s.arrange(func(monitors []placement.Monitor, _ settings.Settings) placement.Monitor {
		return mostOverlapped(monitors, at)
	}, func(size placement.Size, monitors []placement.Monitor, _ settings.Settings) placement.Placed {
		placed, _ := placement.Recover(at, size, monitors)
		return placed
	})
}

// arrange answers the arrangement on the monitor pick chooses, with place deciding the position
// once the size on that monitor is known. Where place lands the strip on another monitor, it is
// sized again in that monitor's pixels and placed again, so a strip moved onto a display at other
// scaling is drawn at that display's size (FR-407).
func (s *Service) arrange(
	pick func([]placement.Monitor, settings.Settings) placement.Monitor,
	place func(placement.Size, []placement.Monitor, settings.Settings) placement.Placed,
) (Arrangement, placement.Monitor, error) {
	monitors, err := s.ports.Monitors.Monitors()
	if err != nil {
		return Arrangement{}, placement.Monitor{}, fmt.Errorf("reading the displays: %w", err)
	}
	if len(monitors) == 0 {
		return Arrangement{}, placement.Monitor{}, ErrNoMonitors
	}
	content := s.stripContent()
	current := content.settings
	monitor := pick(monitors, current)
	size, scrolls := s.stripSize(content, monitor)
	placed := place(size, monitors, current)
	if placed.Monitor.Device != monitor.Device {
		size, scrolls = s.stripSize(content, placed.Monitor)
		placed = place(size, monitors, current)
	}
	arranged := Arrangement{At: placed.At, Size: size, Scrolls: scrolls, DPI: placed.Monitor.DPI}
	return arranged, placed.Monitor, nil
}

// content is what decides the strip's size, read together under one lock.
type content struct {
	settings settings.Settings
	// cells counts every cell the page draws: each notice, then each clock (the prompt standing in
	// for them when there are none).
	cells     int
	scrollbar int
}

func (s *Service) stripContent() content {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	current := s.current.Normalised()
	return content{settings: current, cells: len(s.notices()) + max(len(current.Clocks), 1), scrollbar: s.scrollbar}
}

// stripSize answers the strip's size on monitor in physical pixels (FR-105, FR-106): the cells,
// notices included, fitted along the orientation within the work area; one cell plus padding
// across it, plus the scroll bar's thickness when the cells scroll, so the bar never covers them.
func (s *Service) stripSize(content content, monitor placement.Monitor) (placement.Size, bool) {
	current := content.settings
	cell := s.layout.Digital
	switch {
	case len(current.Clocks) == 0:
		cell = s.layout.Prompt
	case current.Style == settings.Analogue:
		cell = s.layout.Analogue
	}
	along, across := cell.Width, cell.Height
	room := monitor.Work.Width()
	if current.Orientation == settings.Vertical {
		along, across = cell.Height, cell.Width
		room = monitor.Work.Height()
	}
	available := placement.Scale(room, monitor.DPI, placement.BaseDPI)
	fitted := placement.Fit(content.cells, along, s.layout.Padding, available)
	thickness := across + 2*s.layout.Padding
	if fitted.Scrolls {
		thickness += content.scrollbar
	}
	length := placement.Scale(fitted.Length, placement.BaseDPI, monitor.DPI)
	breadth := placement.Scale(thickness, placement.BaseDPI, monitor.DPI)
	if current.Orientation == settings.Vertical {
		return placement.Size{Width: breadth, Height: length}, fitted.Scrolls
	}
	return placement.Size{Width: length, Height: breadth}, fitted.Scrolls
}

// storedOrPrimary answers the stored monitor where it is present; else the primary.
func storedOrPrimary(monitors []placement.Monitor, stored *placement.Stored) placement.Monitor {
	if stored != nil {
		index := slices.IndexFunc(monitors, func(m placement.Monitor) bool { return m.Device == stored.Device })
		if index >= 0 {
			return monitors[index]
		}
	}
	primary, _ := placement.Primary(monitors)
	return primary
}

// mostOverlapped answers the monitor holding at; the primary when none does.
func mostOverlapped(monitors []placement.Monitor, at placement.Point) placement.Monitor {
	placed, _ := placement.Recover(at, placement.Size{Width: 1, Height: 1}, monitors)
	return placed.Monitor
}
