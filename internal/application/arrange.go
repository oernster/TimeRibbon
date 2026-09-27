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

// Launch arranges the strip at launch or as a panel closes: on its stored monitor and offset, else
// at the default place on the primary monitor, always wholly inside a work area (FR-403, FR-405);
// re-centred along its length where that length changed while the panel was open (FR-104).
func (s *Service) Launch() (Arrangement, error) {
	return s.recentredKept(func() (Arrangement, placement.Monitor, bool, error) {
		return s.arrange(func(monitors []placement.Monitor, current settings.Settings) placement.Monitor {
			return storedOrPrimary(monitors, current.Placement)
		}, func(size placement.Size, monitors []placement.Monitor, current settings.Settings) placement.Placed {
			placed, _ := placement.Restore(current.Placement, monitors, size)
			return placed
		})
	})
}

// Rearrange arranges a strip now at at after its content or the displays changed: the same
// top-left corner, moved the least distance that keeps it wholly inside the work area it overlaps
// most (FR-406); re-centred along its length on that work area where its length changed (FR-104).
// Only a re-centring is saved, so a monitor that comes back finds its placement kept.
func (s *Service) Rearrange(at placement.Point) (Arrangement, error) {
	return s.recentredKept(func() (Arrangement, placement.Monitor, bool, error) { return s.recovered(at) })
}

// Moved records where a drag left the strip (FR-404) and answers where it belongs, which differs
// only when the drag left part of it off every work area.
func (s *Service) Moved(at placement.Point) (Arrangement, error) {
	arranged, monitor, _, err := s.recovered(at)
	if err != nil {
		return arranged, err
	}
	return arranged, s.record(arranged.At, monitor)
}

// ToEdge puts a strip now at at flush against edge of the work area it overlaps most, centred along
// that edge; it keeps that place (FR-408).
func (s *Service) ToEdge(at placement.Point, edge placement.Edge) (Arrangement, error) {
	return s.recentredKept(func() (Arrangement, placement.Monitor, bool, error) {
		arranged, monitor, _, err := s.arrange(func(monitors []placement.Monitor, _ settings.Settings) placement.Monitor {
			return mostOverlapped(monitors, at)
		}, func(size placement.Size, monitors []placement.Monitor, _ settings.Settings) placement.Placed {
			monitor := mostOverlapped(monitors, at)
			return placement.Placed{At: placement.AgainstEdge(size, monitor.Work, edge), Monitor: monitor}
		})
		return arranged, monitor, true, err
	})
}

// recentredKept answers what arrange answers, saving the strip's place where arrange says it was
// moved, re-centred or put against an edge, so the next launch finds it there (FR-104, FR-408). A
// save that fails raises a notice, one more cell (FR-707), so the strip is arranged once more to
// fit it; that arrangement is not saved again.
func (s *Service) recentredKept(arrange func() (Arrangement, placement.Monitor, bool, error)) (Arrangement, error) {
	arranged, monitor, recentred, err := arrange()
	if err != nil || !recentred {
		return arranged, err
	}
	if s.record(arranged.At, monitor) == nil {
		return arranged, nil
	}
	arranged, _, _, err = arrange()
	return arranged, err
}

// record stores at as the strip's place on monitor.
func (s *Service) record(at placement.Point, monitor placement.Monitor) error {
	stored := placement.Record(at, monitor)
	return s.change(func(current settings.Settings) (settings.Settings, error) {
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
func (s *Service) recovered(at placement.Point) (Arrangement, placement.Monitor, bool, error) {
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
// scaling is drawn at that display's size (FR-407). Where the strip's length differs from the
// last arrangement's, it is re-centred along its length, which it answers (FR-104).
func (s *Service) arrange(
	pick func([]placement.Monitor, settings.Settings) placement.Monitor,
	place func(placement.Size, []placement.Monitor, settings.Settings) placement.Placed,
) (Arrangement, placement.Monitor, bool, error) {
	monitors, err := s.ports.Monitors.Monitors()
	if err != nil {
		return Arrangement{}, placement.Monitor{}, false, fmt.Errorf("reading the displays: %w", err)
	}
	if len(monitors) == 0 {
		return Arrangement{}, placement.Monitor{}, false, ErrNoMonitors
	}
	content := s.stripContent()
	current := content.settings
	monitor := pick(monitors, current)
	size, scrolls, length := s.stripSize(content, monitor)
	placed := place(size, monitors, current)
	if placed.Monitor.Device != monitor.Device {
		size, scrolls, length = s.stripSize(content, placed.Monitor)
		placed = place(size, monitors, current)
	}
	vertical := current.Orientation == settings.Vertical
	recentred := s.lengthChanged(stripLength{known: true, vertical: vertical, length: length})
	if recentred {
		placed.At = placement.CentredAlong(placed.At, size, placed.Monitor.Work, vertical)
	}
	arranged := Arrangement{At: placed.At, Size: size, Scrolls: scrolls, DPI: placed.Monitor.DPI}
	return arranged, placed.Monitor, recentred, nil
}

// lengthChanged records now as the strip's length, answering whether it differs from the length
// recorded last; never the first time, when there was none.
func (s *Service) lengthChanged(now stripLength) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	changed := s.arranged.known && s.arranged != now
	s.arranged = now
	return changed
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
// It answers the length along the orientation in DIP too, which a move between scalings keeps.
func (s *Service) stripSize(content content, monitor placement.Monitor) (placement.Size, bool, int) {
	current := content.settings
	layout := s.layouts.For(current.Size)
	cell := layout.Digital
	switch {
	case len(current.Clocks) == 0:
		cell = layout.Prompt
	case current.Style == settings.Analogue:
		cell = layout.Analogue
	}
	along, across := cell.Width, cell.Height
	room := monitor.Work.Width()
	if current.Orientation == settings.Vertical {
		along, across = cell.Height, cell.Width
		room = monitor.Work.Height()
	}
	available := placement.Scale(room, monitor.DPI, placement.BaseDPI)
	fitted := placement.Fit(content.cells, along, layout.Padding, available)
	thickness := across + 2*layout.Padding
	if fitted.Scrolls {
		thickness += content.scrollbar
	}
	length := placement.Scale(fitted.Length, placement.BaseDPI, monitor.DPI)
	breadth := placement.Scale(thickness, placement.BaseDPI, monitor.DPI)
	if current.Orientation == settings.Vertical {
		return placement.Size{Width: breadth, Height: length}, fitted.Scrolls, fitted.Length
	}
	return placement.Size{Width: length, Height: breadth}, fitted.Scrolls, fitted.Length
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
