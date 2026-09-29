package main

import (
	"testing"

	"github.com/oernster/timeribbon/internal/application"
	"github.com/oernster/timeribbon/internal/domain/placement"
	"github.com/oernster/timeribbon/internal/infrastructure/desktop"
)

// withMap is testArrange, a ribbon at (10, 20) 300 by 90, with a 480 by 240 map below it centred on
// it: the window runs from x -80 to 400 and y 20 to 350, the ribbon's corner 90 in from its left.
var withMap = func() application.Arrangement {
	arranged := testArrange
	arranged.MapSide = placement.Bottom
	arranged.Map = placement.Rect{Left: -80, Top: 110, Right: 400, Bottom: 350}
	return arranged
}()

// FR-902, FR-909: the window is the ribbon with its map; the ribbon's corner sits inside it.
func TestTheWindowHoldsTheRibbonAndItsMap(t *testing.T) {
	t.Parallel()
	at, size, offset := windowOf(withMap)
	if at != (placement.Point{X: -80, Y: 20}) || size != (placement.Size{Width: 480, Height: 330}) || offset != (placement.Point{X: 90, Y: 0}) {
		t.Errorf("got %+v %+v %+v", at, size, offset)
	}
	if at, size, offset := windowOf(testArrange); at != testArrange.At || size != testArrange.Size || offset != (placement.Point{}) {
		t.Errorf("no map: got %+v %+v %+v", at, size, offset)
	}
}

// FR-909: a drag moves the window; the ribbon's own place is read back through its corner's offset,
// so the drop is decided for the ribbon, never the window.
func TestADragOfTheMapMovesTheRibbonToo(t *testing.T) {
	t.Parallel()
	app, service, seen, _ := newTestApp(t)
	service.arrangement = withMap
	if err := app.placeLaunched(); err != nil {
		t.Fatal(err)
	}
	app.show()
	if got := lastPlaced(t, seen); got.At != (placement.Point{X: -80, Y: 20}) || got.Size != (placement.Size{Width: 480, Height: 330}) {
		t.Fatalf("placed %+v, want the window with its map", got)
	}
	seen.ribbonAt = placement.Point{X: 20, Y: 500}
	app.handleSafely(desktop.Event{Kind: desktop.EventMoveEnded})
	if at := service.at[len(service.at)-1]; at != (placement.Point{X: 110, Y: 500}) {
		t.Errorf("the drop was read as %+v, want the ribbon's corner 90 in", at)
	}
}

// FR-910: the page is told where to draw the map only while the full ribbon shows, never with a tab.
func TestTheMapHidesWithTheTab(t *testing.T) {
	t.Parallel()
	app, service, _ := unpinnedApp(t)
	service.arrangement = withMap
	if err := app.placeLaunched(); err != nil {
		t.Fatal(err)
	}
	if shown := app.Snapshot().SunMap; shown.Shown || !app.Snapshot().Collapsed {
		t.Errorf("collapsed: %+v", shown)
	}
	app.service.(*scriptedService).settings.Pinned = true
	if err := app.placeLaunched(); err != nil {
		t.Fatal(err)
	}
	got := app.Snapshot().SunMap
	if !got.Shown || got.Side != "bottom" || got.Map != (boxDTO{X: 0, Y: 90, Width: 480, Height: 240}) || got.Ribbon != (boxDTO{X: 90, Y: 0, Width: 300, Height: 90}) {
		t.Errorf("shown: %+v", got)
	}
}

// FR-901, FR-903: the menu item and the handle each flip their choice, then refit the window.
func TestTheSunMapItemAndTheHandleFlipTheirChoices(t *testing.T) {
	t.Parallel()
	app, service, _, _ := newTestApp(t)
	app.act(application.ActionSunMap)
	if err := app.TogglePullOut(); err != nil {
		t.Fatal(err)
	}
	if !service.settings.SunMap || !service.settings.PullOut {
		t.Errorf("sun map %v, pull out %v; want both on", service.settings.SunMap, service.settings.PullOut)
	}
	if err := app.TogglePullOut(); err != nil || service.settings.PullOut {
		t.Errorf("the second click left the pull out %v (%v)", service.settings.PullOut, err)
	}
}
