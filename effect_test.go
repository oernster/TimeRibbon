package main

import (
	"slices"
	"testing"

	"github.com/oernster/timeribbon/ribbonkit/application/arranger"
	"github.com/oernster/timeribbon/ribbonkit/application/menus"
	"github.com/oernster/timeribbon/ribbonkit/domain/hover"
	"github.com/oernster/timeribbon/ribbonkit/domain/placement"
	"github.com/oernster/timeribbon/ribbonkit/infrastructure/desktop"
)

// openUnpinnedApp answers an unpinned ribbon opened from its tab and shown in full.
func openUnpinnedApp(t *testing.T) (*App, *scriptedService, *window) {
	t.Helper()
	app, service, seen := unpinnedApp(t)
	app.handleSafely(desktop.Event{Kind: desktop.EventPointerArrived})
	fire(t, seen)
	app.RibbonDrawn()
	if app.Snapshot().Collapsed {
		t.Fatal("the ribbon did not open")
	}
	return app, service, seen
}

// lastOnTop answers the last keeping on top the facade asked for; a failure when it asked for none.
func lastOnTop(t *testing.T, seen *window) bool {
	t.Helper()
	if len(seen.onTop) == 0 {
		t.Fatal("the window was never set on top or off it")
	}
	return seen.onTop[len(seen.onTop)-1]
}

// dropAt ends a drag with the ribbon arranged as arranged.
func dropAt(app *App, service *scriptedService, arranged arranger.Arrangement) {
	service.arrangement = arranged
	app.handleSafely(desktop.Event{Kind: desktop.EventMoveEnded})
}

// FR-619, its acceptance: an unpinned ribbon dragged away from every edge stays in full once the
// pointer has left it, Pin ribbon still unticked; it is no longer kept on top with Always on top off.
func TestAnUnpinnedRibbonOffAnEdgeShowsInFull(t *testing.T) {
	t.Parallel()
	app, service, seen := openUnpinnedApp(t)
	dropAt(app, service, testAway)
	app.handleSafely(desktop.Event{Kind: desktop.EventPointerLeft})
	if seen.pending != nil {
		fire(t, seen)
	}
	if app.Snapshot().Collapsed || lastPlaced(t, seen).Size != testAway.Size {
		t.Errorf("collapsed %v, placed %+v; want the full ribbon", app.Snapshot().Collapsed, lastPlaced(t, seen))
	}
	if service.settings.Pinned || lastOnTop(t, seen) {
		t.Errorf("pinned %v, on top %v; want the choice kept and the ribbon not kept on top", service.settings.Pinned, seen.onTop)
	}
	if seen.tabFrames[len(seen.tabFrames)-1] {
		t.Error("a ribbon pinned in effect wears the tab's frame")
	}
}

// FR-619, its acceptance: dragged back onto an edge, the tab behaviour resumes by itself: the ribbon
// stays in full under the pointer and collapses a second after the pointer leaves.
func TestDraggingBackOntoAnEdgeCollapsesAgain(t *testing.T) {
	t.Parallel()
	app, service, seen := openUnpinnedApp(t)
	dropAt(app, service, testAway)
	app.handleSafely(desktop.Event{Kind: desktop.EventPointerLeft})
	if seen.pending != nil {
		fire(t, seen)
	}
	app.handleSafely(desktop.Event{Kind: desktop.EventPointerArrived})
	dropAt(app, service, testArrange)
	if app.Snapshot().Collapsed || !lastOnTop(t, seen) {
		t.Fatalf("dropped on an edge: collapsed %v, on top %v; want full and on top", app.Snapshot().Collapsed, seen.onTop)
	}
	app.handleSafely(desktop.Event{Kind: desktop.EventPointerLeft})
	if seen.waited != hover.Away {
		t.Fatalf("waiting %v, want the second away", seen.waited)
	}
	fire(t, seen)
	if got := lastPlaced(t, seen); got.Size.Width != placement.TabThickness || !app.Snapshot().Collapsed {
		t.Errorf("placed %+v, want the tab", got)
	}
}

// FR-613, its acceptance: unticking Pin ribbon while the ribbon stands against no edge moves it to the
// edge it last stood against, where it collapses once the pointer is off it.
func TestUnpinningAwayFromAnEdgeMovesItToTheLastEdge(t *testing.T) {
	t.Parallel()
	app, service, seen, _ := newTestApp(t)
	service.arrangement = testAway
	service.lastEdge = &testArrange
	if err := app.placeLaunched(); err != nil {
		t.Fatal(err)
	}
	app.show()
	app.act(menus.Pin)
	if !slices.Contains(service.calls, "ToLastEdge") || lastPlaced(t, seen) != (arranger.Arrangement{At: testArrange.At, Size: testArrange.Size}) {
		t.Fatalf("calls %v, placed %+v; want the last edge", service.calls, lastPlaced(t, seen))
	}
	if seen.waited != hover.Away {
		t.Fatalf("waiting %v, want the second away", seen.waited)
	}
	fire(t, seen)
	if !app.Snapshot().Collapsed {
		t.Error("the ribbon did not collapse on its edge")
	}
}

// FR-613: unticking Pin ribbon while flush moves nothing; a Position choice keeps the chosen pin.
func TestUnpinningOnAnEdgeMovesNothingAndRecentringKeepsThePin(t *testing.T) {
	t.Parallel()
	app, service, _, _ := newTestApp(t)
	if err := app.placeLaunched(); err != nil {
		t.Fatal(err)
	}
	app.show()
	app.act(menus.Pin)
	if slices.Contains(service.calls, "ToLastEdge") {
		t.Error("unpinning a flush ribbon moved it")
	}
	app.toEdge(placement.Left)
	if service.settings.Pinned || !slices.Contains(service.calls, "ToEdge") {
		t.Errorf("after a Position choice: pinned %v, calls %v", service.settings.Pinned, service.calls)
	}
}
