package application

import (
	"errors"
	"testing"

	"github.com/oernster/timestrip/internal/domain/placement"
	"github.com/oernster/timestrip/internal/domain/settings"
)

// clocks answers settings holding n London clocks.
func clocks(n int) settings.Settings {
	s := settings.Defaults()
	for range n {
		s = s.WithClockAdded(settings.Entry{ID: "x", Zone: "Europe/London"})
	}
	return s
}

// Two digital cells at 100 percent: 2 x 160 + 2 x 8 = 336 along, 90 + 2 x 8 = 106 across.
func TestLaunchWithNothingStoredGoesToTheDefaultPlace(t *testing.T) {
	t.Parallel()
	r := newRig(t, clocks(2))
	got, err := r.service.Launch()
	if err != nil {
		t.Fatal(err)
	}
	want := Arrangement{
		At:   placement.Point{X: 1920 - placement.EdgeMarginDIP - 336, Y: (1032 - 106) / 2},
		Size: placement.Size{Width: 336, Height: 106}, DPI: placement.BaseDPI,
	}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

// FR-405, FR-407: the stored monitor at 150 percent sizes the strip in its pixels.
func TestLaunchRestoresTheStoredMonitorAtItsScaling(t *testing.T) {
	t.Parallel()
	initial := clocks(2)
	initial.Placement = &placement.Stored{Device: secondaryMonitor.Device, DPI: 144, Offset: placement.Point{X: 180, Y: 300}}
	r := newRig(t, initial)
	got, err := r.service.Launch()
	if err != nil {
		t.Fatal(err)
	}
	want := Arrangement{At: placement.Point{X: 2100, Y: 300}, Size: placement.Size{Width: 504, Height: 159}, DPI: 144}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

// FR-103: three analogue cells stacked: 160 + 16 across, 3 x 150 + 16 along.
func TestVerticalStripsStackTheirCells(t *testing.T) {
	t.Parallel()
	initial := clocks(3)
	initial.Orientation = settings.Vertical
	initial.Style = settings.Analogue
	r := newRig(t, initial)
	got, _ := r.service.Launch()
	if got.Size != (placement.Size{Width: 176, Height: 466}) || got.Scrolls {
		t.Errorf("got %+v", got)
	}
}

// FR-106.
func TestAStripThatWillNotFitScrollsAtTheWidthOfTheWorkArea(t *testing.T) {
	t.Parallel()
	r := newRig(t, clocks(12))
	got, _ := r.service.Launch()
	if got.Size.Width != 1920 || !got.Scrolls || got.At.X != 0 {
		t.Errorf("got %+v", got)
	}
}

// FR-404.
func TestPlacementIsStoredRelativeToItsMonitor(t *testing.T) {
	t.Parallel()
	r := newRig(t, clocks(2))
	got, err := r.service.Moved(placement.Point{X: 2100, Y: 300})
	if err != nil {
		t.Fatal(err)
	}
	stored := r.store.last(t).Placement
	want := placement.Stored{Device: secondaryMonitor.Device, Work: secondaryMonitor.Work, DPI: 144, Offset: placement.Point{X: 180, Y: 300}}
	if stored == nil || *stored != want || got.At != (placement.Point{X: 2100, Y: 300}) {
		t.Errorf("stored %+v at %+v", stored, got.At)
	}
}

// FR-406: a strip dragged off every display comes back; the place it comes back to is stored.
func TestADragOffEveryDisplayIsBroughtBack(t *testing.T) {
	t.Parallel()
	r := newRig(t, clocks(2))
	got, err := r.service.Moved(placement.Point{X: 9000, Y: 300})
	if err != nil {
		t.Fatal(err)
	}
	if got.At != (placement.Point{X: 1568, Y: 463}) || r.store.last(t).Placement.Device != primaryMonitor.Device {
		t.Errorf("got %+v, stored %+v", got, r.store.last(t).Placement)
	}
}

// FR-104, FR-406: rearranging keeps the corner, clamps it and saves nothing.
func TestRearrangingClampsAndSavesNothing(t *testing.T) {
	t.Parallel()
	r := newRig(t, clocks(2))
	got, err := r.service.Rearrange(placement.Point{X: 1500, Y: 1000})
	if err != nil {
		t.Fatal(err)
	}
	if got.At != (placement.Point{X: 1500, Y: 1032 - 106}) {
		t.Errorf("got %+v", got.At)
	}
	if len(r.store.saved) != 0 {
		t.Error("rearranging saved a placement")
	}
}

// FR-407: at (1800, 1000) the strip overlaps the secondary most, so it is sized in the secondary's
// pixels (504 by 159 at 150 percent) and clamped onto it.
func TestAStripLandingOnAnotherDisplayIsSizedForIt(t *testing.T) {
	t.Parallel()
	r := newRig(t, clocks(2))
	got, err := r.service.Rearrange(placement.Point{X: 1800, Y: 1000})
	if err != nil {
		t.Fatal(err)
	}
	want := Arrangement{At: placement.Point{X: 1920, Y: 1000}, Size: placement.Size{Width: 504, Height: 159}, DPI: 144}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

// CON-6: Settings opens centred on the strip's display, sized in its pixels.
func TestSettingsOpenCentredOnTheStripsDisplay(t *testing.T) {
	t.Parallel()
	r := newRig(t, clocks(1))
	got, err := r.service.Centred(placement.Point{X: 2100, Y: 300}, placement.Size{Width: 400, Height: 300})
	if err != nil {
		t.Fatal(err)
	}
	want := Arrangement{At: placement.Point{X: 1920 + (2560-600)/2, Y: (1392 - 450) / 2}, Size: placement.Size{Width: 600, Height: 450}, DPI: 144}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
	if len(r.store.saved) != 0 {
		t.Error("centring saved something")
	}
}

func TestNoDisplaysOrAFaultReadingThemIsAnswered(t *testing.T) {
	t.Parallel()
	r := newRig(t, clocks(1))
	r.service.ports.Monitors = fakeMonitors{}
	if _, err := r.service.Launch(); !errors.Is(err, ErrNoMonitors) {
		t.Errorf("no monitors: got %v", err)
	}
	r.service.ports.Monitors = fakeMonitors{err: errPlanted}
	if _, err := r.service.Rearrange(placement.Point{}); !errors.Is(err, errPlanted) {
		t.Errorf("a fault: got %v", err)
	}
	if _, err := r.service.Moved(placement.Point{}); !errors.Is(err, errPlanted) {
		t.Errorf("a fault while moving: got %v", err)
	}
	if _, err := r.service.Centred(placement.Point{}, placement.Size{}); !errors.Is(err, errPlanted) {
		t.Errorf("a fault while centring: got %v", err)
	}
	r.service.ports.Monitors = fakeMonitors{}
	if _, err := r.service.Centred(placement.Point{}, placement.Size{}); !errors.Is(err, ErrNoMonitors) {
		t.Errorf("centring with no monitors: got %v", err)
	}
}
