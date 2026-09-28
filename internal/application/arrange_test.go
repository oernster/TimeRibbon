package application

import (
	"errors"
	"testing"

	"github.com/oernster/timeribbon/internal/domain/placement"
	"github.com/oernster/timeribbon/internal/domain/settings"
)

// clocks answers horizontal settings holding n London clocks; the arithmetic below is worked for
// horizontal cells, whatever the default orientation is.
func clocks(n int) settings.Settings {
	s := settings.Defaults()
	s.Orientation = settings.Horizontal
	for range n {
		s = s.WithClockAdded(settings.Entry{ID: "x", Zone: "Europe/London"})
	}
	return s
}

// FR-403, FR-409: two digital cells at 100 percent, 2 x 160 + 2 x 8 = 336 along and
// 90 + 2 x 8 = 106 across, go to their orientation's home edge: flush against the top, centred left
// to right, when horizontal; flush against the right, centred top to bottom, when vertical.
func TestLaunchWithNothingStoredGoesToTheDefaultPlace(t *testing.T) {
	t.Parallel()
	r := newRig(t, clocks(2))
	got, err := r.service.Launch()
	if err != nil {
		t.Fatal(err)
	}
	want := Arrangement{
		At:   placement.Point{X: (1920 - 336) / 2, Y: 0},
		Size: placement.Size{Width: 336, Height: 106}, DPI: placement.BaseDPI,
	}
	if got != want {
		t.Errorf("horizontal: got %+v, want %+v", got, want)
	}
	vertical := clocks(2)
	vertical.Orientation = settings.Vertical
	// Vertical, the same two cells are 160 + 2 x 8 = 176 across and 2 x 90 + 2 x 8 = 196 along.
	if got, _ := newRig(t, vertical).service.Launch(); got.At != (placement.Point{X: 1920 - 176, Y: (1032 - 196) / 2}) {
		t.Errorf("vertical: got %+v", got)
	}
}

// FR-405, FR-407: the stored monitor at 150 percent sizes the ribbon in its pixels.
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
func TestVerticalRibbonsStackTheirCells(t *testing.T) {
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

// draggedTo answers settings holding n London clocks in orientation, stored where a drag left the
// ribbon at at on the primary monitor.
func draggedTo(n int, orientation settings.Orientation, at placement.Point) settings.Settings {
	s := clocks(n)
	s.Orientation = orientation
	s.Placement = &placement.Stored{Device: primaryMonitor.Device, DPI: placement.BaseDPI, Offset: at}
	return s
}

// FR-104: a vertical ribbon dragged near the top gains a clock. Its length changes from
// 2 x 90 + 8 = 188 to 3 x 90 + 2 x 8 = 286, so it is centred top to bottom, (1032 - 286) / 2,
// its left edge kept; the place is saved, so the next launch finds it there.
func TestARibbonWhoseLengthChangesIsRecentredAndKept(t *testing.T) {
	t.Parallel()
	dragged := placement.Point{X: 1700, Y: 40}
	r := newRig(t, draggedTo(2, settings.Vertical, dragged))
	if got, _ := r.service.Launch(); got.At != dragged {
		t.Fatalf("launched at %+v, want where the drag left it", got.At)
	}
	if _, err := r.service.AddClock("Asia/Kolkata"); err != nil {
		t.Fatal(err)
	}
	centred := placement.Point{X: 1700, Y: (1032 - 286) / 2}
	got, err := r.service.Rearrange(dragged)
	if err != nil || got.At != centred {
		t.Errorf("rearranged to %+v (%v), want %+v", got.At, err, centred)
	}
	if stored := r.store.last(t).Placement; stored == nil || stored.Offset != centred {
		t.Errorf("stored %+v, want the centred place", stored)
	}
	if got, _ := r.service.Launch(); got.At != centred {
		t.Errorf("relaunched at %+v, want the centred place", got.At)
	}
}

// FR-104: a horizontal ribbon is centred left to right, its top kept.
func TestAHorizontalRibbonIsRecentredLeftToRight(t *testing.T) {
	t.Parallel()
	dragged := placement.Point{X: 30, Y: 800}
	r := newRig(t, draggedTo(2, settings.Horizontal, dragged))
	_, _ = r.service.Launch()
	_, _ = r.service.AddClock("Asia/Kolkata")
	got, _ := r.service.Rearrange(dragged)
	if want := (placement.Point{X: (1920 - (3*160 + 2*8)) / 2, Y: 800}); got.At != want {
		t.Errorf("got %+v, want %+v", got.At, want)
	}
}

// FR-104, FR-404: only a change of length re-centres the ribbon; a rearrange or a move with the
// same clocks leaves it where it was put. Nothing is saved but the move.
func TestNothingButAChangeOfLengthRecentresTheRibbon(t *testing.T) {
	t.Parallel()
	dragged := placement.Point{X: 1700, Y: 40}
	r := newRig(t, draggedTo(2, settings.Vertical, dragged))
	_, _ = r.service.Launch()
	if got, _ := r.service.Rearrange(dragged); got.At != dragged || len(r.store.saved) != 0 {
		t.Errorf("rearranged to %+v with %d saves, want it left alone", got.At, len(r.store.saved))
	}
	moved := placement.Point{X: 1700, Y: 500}
	if got, _ := r.service.Moved(moved); got.At != moved || r.store.last(t).Placement.Offset != moved {
		t.Errorf("a drag was moved to %+v, want it kept where it was let go", got.At)
	}
}

// FR-107: an empty ribbon is sized for its prompt, whatever the style: 160 + 16 by 190 + 16.
func TestAnEmptyRibbonIsSizedForItsPrompt(t *testing.T) {
	t.Parallel()
	r := newRig(t, clocks(0))
	got, _ := r.service.Launch()
	if got.Size != (placement.Size{Width: 176, Height: 206}) {
		t.Errorf("got %+v", got.Size)
	}
}

// FR-106.
func TestARibbonThatWillNotFitScrollsAtTheWidthOfTheWorkArea(t *testing.T) {
	t.Parallel()
	r := newRig(t, clocks(12))
	got, _ := r.service.Launch()
	if got.Size.Width != 1920 || !got.Scrolls || got.At.X != 0 {
		t.Errorf("got %+v", got)
	}
}

// FR-106, FR-707: a notice is drawn as one more cell, so the ribbon makes room for it; two clocks and
// a notice are 3 x 160 + 2 x 8 = 496 along. Once the notice is dismissed the ribbon is 336 again.
func TestTheRibbonMakesRoomForANotice(t *testing.T) {
	t.Parallel()
	r := newRig(t, clocks(2))
	r.store.saveErr = errPlanted
	if err := r.service.SetTheme(settings.Dark); !errors.Is(err, errPlanted) {
		t.Fatalf("the save did not fail: %v", err)
	}
	got, _ := r.service.Launch()
	if got.Size.Width != 496 || got.Scrolls {
		t.Errorf("with a notice: got %+v", got)
	}
	r.store.saveErr = nil
	r.service.DismissNotices()
	got, _ = r.service.Launch()
	if got.Size.Width != 336 {
		t.Errorf("after dismissing: got %+v", got)
	}
}

// FR-104, FR-707: a ribbon re-centred where its place cannot be saved raises the notice again, so it
// is arranged once more with room for that cell rather than left too short for it.
func TestARecentringThatCannotBeSavedMakesRoomForItsNotice(t *testing.T) {
	t.Parallel()
	r := newRig(t, clocks(2))
	r.store.saveErr = errPlanted
	_ = r.service.SetTheme(settings.Dark)
	if _, err := r.service.Launch(); err != nil {
		t.Fatal(err)
	}
	r.service.DismissNotices()
	got, err := r.service.Launch()
	if err != nil || got.Size.Width != 496 {
		t.Errorf("got %+v (%v), want room for the notice the failed save raised", got, err)
	}
	if notices := r.service.Snapshot().Notices; len(notices) != 1 {
		t.Errorf("notices %v, want the failed save's", notices)
	}
}

// FR-106: a ribbon that scrolls is made thicker by the scroll bar the page reports, so the bar never
// covers the cells; one that fits is not. 12 cells overflow the primary: 90 + 16 + 15 = 121 across.
func TestAScrollingRibbonMakesRoomForItsScrollBar(t *testing.T) {
	t.Parallel()
	r := newRig(t, clocks(12))
	const bar = 15
	if err := r.service.SetScrollbar(bar); err != nil {
		t.Fatal(err)
	}
	got, _ := r.service.Launch()
	if !got.Scrolls || got.Size.Height != 90+2*testLayout.Padding+bar {
		t.Errorf("scrolling: got %+v", got)
	}
	fits := newRig(t, clocks(2))
	if err := fits.service.SetScrollbar(bar); err != nil {
		t.Fatal(err)
	}
	if got, _ := fits.service.Launch(); got.Size.Height != 90+2*testLayout.Padding {
		t.Errorf("fitting: got %+v", got)
	}
	if err := r.service.SetScrollbar(-1); !errors.Is(err, ErrNegativeLength) {
		t.Errorf("a negative bar: got %v", err)
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

// FR-406: a horizontal ribbon dragged off every display comes back to the default place, flush
// against the primary's top (FR-403, FR-409), 336 x 106; the place it comes back to is stored.
func TestADragOffEveryDisplayIsBroughtBack(t *testing.T) {
	t.Parallel()
	r := newRig(t, clocks(2))
	got, err := r.service.Moved(placement.Point{X: 9000, Y: 300})
	if err != nil {
		t.Fatal(err)
	}
	if got.At != (placement.Point{X: (1920 - 336) / 2, Y: 0}) || r.store.last(t).Placement.Device != primaryMonitor.Device {
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

// FR-407: at (1800, 1000) the ribbon overlaps the secondary most, so it is sized in the secondary's
// pixels (504 by 159 at 150 percent) and clamped onto it.
func TestARibbonLandingOnAnotherDisplayIsSizedForIt(t *testing.T) {
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

// CON-6: Settings opens centred on the ribbon's display, sized in its pixels.
func TestSettingsOpenCentredOnTheRibbonsDisplay(t *testing.T) {
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
	tall, _ := r.service.Centred(placement.Point{X: 10, Y: 10}, placement.Size{Width: 400, Height: 2000})
	if tall.Size.Height != 1032 || tall.At.Y != 0 {
		t.Errorf("a surface taller than the work area is not capped to it: %+v", tall)
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
