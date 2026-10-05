package application

import (
	"errors"
	"testing"

	"github.com/oernster/timeribbon/internal/domain/settings"
	"github.com/oernster/timeribbon/ribbonkit/domain/placement"
)

// FR-614: the tab is the band on the side flush against the ribbon's edge, in that display's pixels.
// Collapsing saves nothing, so the stored place stays the full ribbon's.
func TestCollapsingKeepsThePlacement(t *testing.T) {
	t.Parallel()
	r := newRig(t, settings.Defaults())
	full, err := r.service.Launch()
	if err != nil {
		t.Fatal(err)
	}
	saves := len(r.store.saved)
	tab, err := r.service.Collapsed(full)
	if err != nil {
		t.Fatal(err)
	}
	want := Arrangement{
		At:   placement.Point{X: primaryMonitor.Work.Right - placement.TabThickness, Y: full.At.Y},
		Size: placement.Size{Width: placement.TabThickness, Height: full.Size.Height},
		DPI:  placement.BaseDPI,
	}
	if tab != want {
		t.Errorf("got %+v, want %+v", tab, want)
	}
	if len(r.store.saved) != saves {
		t.Error("collapsing saved the settings")
	}
	// At 150 percent the band is 12 pixels deep, on the side against that display's right edge.
	wide := Arrangement{At: placement.Point{X: 4000, Y: 100}, Size: placement.Size{Width: 264, Height: 300}, Edge: placement.Right}
	scaled, err := r.service.Collapsed(wide)
	if err != nil {
		t.Fatal(err)
	}
	if scaled.At.X != 4264-12 || scaled.Size.Width != 12 || scaled.DPI != secondaryMonitor.DPI {
		t.Errorf("at 150 percent: got %+v", scaled)
	}
}

func TestCollapsingWithoutDisplaysIsRefused(t *testing.T) {
	t.Parallel()
	r := newRig(t, settings.Defaults())
	flush := Arrangement{Edge: placement.Right}
	r.service.ports.Monitors = fakeMonitors{}
	if _, err := r.service.Collapsed(flush); !errors.Is(err, ErrNoMonitors) {
		t.Errorf("no monitors: got %v", err)
	}
	r.service.ports.Monitors = fakeMonitors{err: errPlanted}
	if _, err := r.service.Collapsed(flush); !errors.Is(err, errPlanted) {
		t.Errorf("a fault: got %v", err)
	}
}

// FR-619: a ribbon flush against no edge never collapses, so its tab is refused.
func TestARibbonAgainstNoEdgeHasNoTab(t *testing.T) {
	t.Parallel()
	r := newRig(t, settings.Defaults())
	if _, err := r.service.Collapsed(Arrangement{At: placement.Point{X: 800, Y: 300}}); !errors.Is(err, ErrNotAgainstAnEdge) {
		t.Errorf("got %v", err)
	}
}

// FR-613: both menus hold Pin ribbon directly after Always on top, ticked while pinned; the choice
// is kept.
func TestBothMenusOfferPinAfterAlwaysOnTop(t *testing.T) {
	t.Parallel()
	r := newRig(t, settings.Defaults())
	for name, menu := range map[string][]MenuItem{"tray": r.service.TrayMenu(true), "context": r.service.ContextMenu()} {
		index := -1
		for position, item := range menu {
			if item.Action == ActionAlwaysOnTop {
				index = position
			}
		}
		if index < 0 || index+1 >= len(menu) {
			t.Fatalf("%s: no item follows Always on top", name)
		}
		if pin := menu[index+1]; pin.Action != ActionPin || pin.Label != labelPin || !pin.Checkable || !pin.Checked {
			t.Errorf("%s: after Always on top came %+v, want Pin ribbon ticked", name, pin)
		}
	}
	if err := r.service.SetPinned(false); err != nil {
		t.Fatal(err)
	}
	if find(t, r.service.ContextMenu(), labelPin).Checked {
		t.Error("Pin ribbon is still ticked once unpinned")
	}
	if saved := r.store.saved[len(r.store.saved)-1]; saved.Pinned {
		t.Error("unpinning was not saved")
	}
}
