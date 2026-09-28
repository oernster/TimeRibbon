package desktop

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/oernster/timeribbon/internal/domain/placement"
	"github.com/oernster/timeribbon/internal/infrastructure/gtkmain"
	"github.com/oernster/timeribbon/internal/infrastructure/monitors"
)

func TestMain(m *testing.M) { os.Exit(gtkmain.ServeTests(m.Run)) }

// The window manager acts on a placement when it gets to it, so the result is polled for.
const (
	settleLimit = 2 * time.Second
	settlePause = 20 * time.Millisecond
)

// The test ribbon's size, in GTK's units: long and short, as a horizontal ribbon is.
var testSize = placement.Size{Width: 300, Height: 80}

func TestTheRibbonIsFoundByItsTitle(t *testing.T) {
	made := newTestWindow()
	defer closeTestWindow(made)
	found, err := FindRibbon("")
	if err != nil || found != made {
		t.Errorf("found %d (%v), made %d", found, err, made)
	}
}

// FR-405 on X11: the ribbon stands where it is placed, at the size it is given. This is the
// measurement the Linux design rests on.
func TestTheRibbonGoesWhereItIsPlaced(t *testing.T) {
	ribbon := newTestWindow()
	defer closeTestWindow(ribbon)
	work := primaryWorkArea(t)
	target := placement.Point{X: work.Left + work.Width()/4, Y: work.Top + work.Height()/4}
	if err := Place(ribbon, target, testSize); err != nil {
		t.Fatal(err)
	}
	var at placement.Point
	var got placement.Size
	for deadline := time.Now().Add(settleLimit); time.Now().Before(deadline); time.Sleep(settlePause) {
		at, _ = Position(ribbon)
		got, _ = size(ribbon)
		if at == target && got == testSize {
			break
		}
	}
	t.Logf("placed at %+v size %+v; stands at %+v size %+v", target, testSize, at, got)
	if at != target || got != testSize {
		t.Errorf("the ribbon stands at %+v size %+v, not where it was placed", at, got)
	}
}

// FR-101.
func TestTheRibbonIsKeptOffTheTaskbar(t *testing.T) {
	ribbon := newTestWindow()
	defer closeTestWindow(ribbon)
	if err := HideFromTaskbar(ribbon); err != nil {
		t.Fatal(err)
	}
	if skips, err := skipsTaskbar(ribbon); err != nil || !skips {
		t.Errorf("skips the taskbar %v (%v)", skips, err)
	}
}

// FR-401: the threshold is the desktop's own, which is never zero.
func TestTheDragThresholdIsTheDesktopsOwn(t *testing.T) {
	got := DragThreshold()
	if got.Width <= 0 || got.Width != got.Height {
		t.Errorf("got %+v", got)
	}
	t.Logf("drag threshold %+v DIP", got)
}

// A Window this package never handed out is refused rather than acted on.
func TestAnUnknownWindowIsRefused(t *testing.T) {
	if err := Place(Window(^uintptr(0)), placement.Point{}, testSize); !errors.Is(err, errUnknownWindow) {
		t.Errorf("got %v", err)
	}
}

// primaryWorkArea answers the primary display's work area.
func primaryWorkArea(t *testing.T) placement.Rect {
	t.Helper()
	found, err := monitors.Monitors{}.Monitors()
	if err != nil {
		t.Fatal(err)
	}
	for _, monitor := range found {
		if monitor.Primary {
			return monitor.Work
		}
	}
	t.Fatal("no primary display")
	return placement.Rect{}
}
