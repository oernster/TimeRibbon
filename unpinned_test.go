package main

import (
	"strings"
	"testing"

	"github.com/oernster/timeribbon/internal/application"
	"github.com/oernster/timeribbon/internal/domain/hover"
	"github.com/oernster/timeribbon/internal/domain/placement"
	"github.com/oernster/timeribbon/internal/infrastructure/desktop"
)

// unpinnedApp answers a facade whose ribbon is unpinned, launched and shown: its tab.
func unpinnedApp(t *testing.T) (*App, *scriptedService, *window) {
	t.Helper()
	app, service, seen, _ := newTestApp(t)
	service.settings.Pinned = false
	if err := app.placeLaunched(); err != nil {
		t.Fatal(err)
	}
	app.show()
	return app, service, seen
}

// fire makes the pending timer fall due, as time.AfterFunc would once its wait was over.
func fire(t *testing.T, seen *window) {
	t.Helper()
	do := seen.pending
	if do == nil {
		t.Fatal("nothing is pending")
	}
	seen.pending = nil
	seen.now = seen.now.Add(seen.waited)
	do()
}

func lastPlaced(t *testing.T, seen *window) application.Arrangement {
	t.Helper()
	if len(seen.placed) == 0 {
		t.Fatal("nothing was placed")
	}
	return seen.placed[len(seen.placed)-1]
}

// FR-614: launched unpinned, the window is the tab framed for it; the page is told to draw it.
func TestLaunchingUnpinnedShowsTheTab(t *testing.T) {
	t.Parallel()
	app, _, seen := unpinnedApp(t)
	if got := lastPlaced(t, seen); got.Size.Width != placement.TabThickness || got.At.X != testArrange.At.X+testArrange.Size.Width-placement.TabThickness {
		t.Errorf("placed %+v, want the tab", got)
	}
	if len(seen.tabFrames) == 0 || !seen.tabFrames[len(seen.tabFrames)-1] {
		t.Errorf("frames %v, want the tab's last", seen.tabFrames)
	}
	if !app.Snapshot().Collapsed {
		t.Error("the snapshot does not say the ribbon is collapsed")
	}
	if len(seen.watching) == 0 || !seen.watching[len(seen.watching)-1] {
		t.Error("the pointer is not watched while the tab is shown")
	}
}

// FR-615, FR-616: resting on the tab opens the full ribbon; leaving collapses it a second later.
func TestTheTabOpensAfterTheRestAndCollapsesOnceAway(t *testing.T) {
	t.Parallel()
	app, _, seen := unpinnedApp(t)
	app.handleSafely(desktop.Event{Kind: desktop.EventPointerArrived})
	if seen.waited != hover.Rest {
		t.Fatalf("waiting %v, want the rest", seen.waited)
	}
	fire(t, seen)
	if got := lastPlaced(t, seen); got != (application.Arrangement{At: testArrange.At, Size: testArrange.Size}) {
		t.Errorf("opened at %+v, want the full ribbon", got)
	}
	if app.Snapshot().Collapsed || !seen.sawEvent(eventRefresh) {
		t.Error("the page was not told the ribbon opened")
	}
	app.handleSafely(desktop.Event{Kind: desktop.EventPointerLeft})
	if seen.waited != hover.Away {
		t.Fatalf("waiting %v, want the second away", seen.waited)
	}
	fire(t, seen)
	if got := lastPlaced(t, seen); got.Size.Width != placement.TabThickness {
		t.Errorf("collapsed to %+v, want the tab", got)
	}
}

// FR-616: a panel and the ribbon's own menu each hold it open; it collapses once both have ended.
func TestAPanelAndTheMenuHoldTheRibbonOpen(t *testing.T) {
	t.Parallel()
	app, _, seen := unpinnedApp(t)
	app.ShowContextMenu()
	if err := app.OpenPanel(); err != nil {
		t.Fatal(err)
	}
	app.handleSafely(desktop.Event{Kind: desktop.EventPointerLeft})
	app.handleSafely(desktop.Event{Kind: desktop.EventMenuClosed})
	if seen.pending != nil {
		t.Fatal("the ribbon would collapse while its panel stands")
	}
	if err := app.ClosePanel(); err != nil {
		t.Fatal(err)
	}
	if got := lastPlaced(t, seen); got.Size != testArrange.Size {
		t.Errorf("the panel closed onto %+v, want the full ribbon", got)
	}
	if seen.pending == nil || seen.waited != hover.Away {
		t.Errorf("after the panel closed: waiting %v, want the second away", seen.waited)
	}
}

// FR-613, FR-617: pinning shows the ribbon in full and stops watching the pointer; unpinning keeps it
// on top whatever Always on top holds.
func TestPinningAndUnpinning(t *testing.T) {
	t.Parallel()
	app, service, seen := unpinnedApp(t)
	app.act(application.ActionPin)
	if !service.settings.Pinned || lastPlaced(t, seen).Size != testArrange.Size {
		t.Errorf("pinning left pinned %v, placed %+v", service.settings.Pinned, lastPlaced(t, seen))
	}
	if seen.watching[len(seen.watching)-1] || seen.onTop[len(seen.onTop)-1] {
		t.Error("pinned, the pointer is still watched or the ribbon kept on top")
	}
	app.handleSafely(desktop.Event{Kind: desktop.EventPointerLeft})
	if seen.pending != nil {
		t.Error("a pinned ribbon heard the pointer")
	}
	app.act(application.ActionPin)
	if service.settings.Pinned || !seen.onTop[len(seen.onTop)-1] || seen.waited != hover.Away {
		t.Errorf("unpinning: pinned %v, on top %v, waiting %v", service.settings.Pinned, seen.onTop, seen.waited)
	}
}

// FR-618: a collapsed ribbon counts as shown, so a second launch hides it, tab included.
func TestASecondLaunchHidesACollapsedRibbon(t *testing.T) {
	t.Parallel()
	app, _, seen := unpinnedApp(t)
	app.secondInstance()
	if seen.hidden != 1 || seen.watching[len(seen.watching)-1] {
		t.Errorf("hidden %d, watching %v; want hidden with the pointer unwatched", seen.hidden, seen.watching)
	}
}

// A panic while opening or collapsing, on the timer's own goroutine, is logged, not fatal.
func TestAFailureWhileOpeningIsLogged(t *testing.T) {
	t.Parallel()
	app, _, seen, log := newTestApp(t)
	app.service.(*scriptedService).settings.Pinned = false
	app.handleSafely(desktop.Event{Kind: desktop.EventPointerArrived})
	app.handleSafely(desktop.Event{Kind: desktop.EventPointerLeft})
	app.handleSafely(desktop.Event{Kind: desktop.EventPointerArrived})
	app.place = func(placement.Point, placement.Size) error { panic("planted") }
	app.unpin.shownOpen = false
	fire(t, seen)
	if !strings.Contains(log.String(), "recovered from planted") {
		t.Errorf("log %q", log.String())
	}
}
