package main

import (
	"slices"
	"strings"
	"testing"

	"github.com/oernster/timeribbon/internal/application"
	"github.com/oernster/timeribbon/ribbonkit/application/arranger"
	"github.com/oernster/timeribbon/ribbonkit/domain/placement"
	"github.com/oernster/timeribbon/ribbonkit/infrastructure/desktop"
)

// withMap is testArrange, a ribbon at (10, 20) 300 by 90, with a 480 by 240 map below it centred on
// it: the window runs from x -80 to 400 and y 20 to 350, the ribbon's corner 90 in from its left.
var withMap = func() arranger.Arrangement {
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

// FR-615, FR-903: the page is told where the ribbon and its map go in its own units, the window
// pixels divided by the pixels to each unit that windows are sized with, so it draws them at their
// size even while the window is still the tab. A ratio the service refused changes nothing.
func TestTheMapsPartsReachThePageInItsOwnUnits(t *testing.T) {
	t.Parallel()
	app, service, _ := unpinnedApp(t)
	service.arrangement = withMap
	service.settings.Pinned = true
	const ratio = 1.25
	if err := app.SetPixelRatio(ratio); err != nil {
		t.Fatal(err)
	}
	if err := app.placeLaunched(); err != nil {
		t.Fatal(err)
	}
	perDIP := desktop.PixelsPerDIP(ratio, testUnscaled)
	want := boxDTO{X: 90 / perDIP, Y: 0, Width: 300 / perDIP, Height: 90 / perDIP}
	if got := app.Snapshot().SunMap; got.Ribbon != want || got.Map.Width != 480/perDIP {
		t.Errorf("at %v pixels to a unit the page was told %+v, want the ribbon at %+v", perDIP, got, want)
	}
	service.changeErr = errPlanted
	_ = app.SetPixelRatio(2 * ratio)
	if got := app.Snapshot().SunMap.Ribbon; got != want {
		t.Errorf("a refused ratio moved the ribbon to %+v", got)
	}
}

// FR-615, FR-910: an opening ribbon is drawn before the window grows, so the page is told of its map
// while it draws, not only once grown. Measured 2026-09-29: told of no map, the page drew the ribbon
// alone and the window then grew round a blank map.
func TestAnOpeningRibbonIsDrawnWithItsMap(t *testing.T) {
	t.Parallel()
	app, service, seen, _ := newTestApp(t)
	service.settings.Pinned = false
	service.arrangement = withMap
	if err := app.placeLaunched(); err != nil {
		t.Fatal(err)
	}
	app.show()
	app.handleSafely(desktop.Event{Kind: desktop.EventPointerArrived})
	fire(t, seen)
	if seen.drawPending == nil {
		t.Fatal("opening did not wait for the page")
	}
	got := app.Snapshot()
	if got.Collapsed || !got.SunMap.Shown || got.SunMap.Map != (boxDTO{X: 0, Y: 90, Width: 480, Height: 240}) {
		t.Errorf("while drawing the page was told %+v, collapsed %v; want the ribbon with its map", got.SunMap, got.Collapsed)
	}
}

// FR-913: every placing of the window is cut first, to the ribbon and its map while the map shows and
// to the whole window otherwise: no map, the tab, a panel. A cut that fails still places the window.
func TestTheShapeFollowsEveryRefit(t *testing.T) {
	t.Parallel()
	app, service, seen, log := newTestApp(t)
	service.arrangement = withMap
	app.place = func(at placement.Point, size placement.Size) error {
		if len(seen.shapes) != len(seen.placed)+1 {
			t.Errorf("placed after %d cuts, want the cut first", len(seen.shapes))
		}
		seen.placed = append(seen.placed, arranger.Arrangement{At: at, Size: size})
		return nil
	}
	lastShape := func() []placement.Rect { return seen.shapes[len(seen.shapes)-1] }
	whole := func(size placement.Size) []placement.Rect {
		return []placement.Rect{{Right: size.Width, Bottom: size.Height}}
	}
	if err := app.placeLaunched(); err != nil {
		t.Fatal(err)
	}
	ribbon := placement.Rect{Left: 90, Top: 0, Right: 390, Bottom: 90}
	sunMap := placement.Rect{Left: 0, Top: 90, Right: 480, Bottom: 330}
	if got := lastShape(); len(got) != 2 || got[0] != ribbon || got[1] != sunMap {
		t.Errorf("with the map: cut to %+v", got)
	}
	service.arrangement = testArrange
	if err := app.placeLaunched(); err != nil {
		t.Fatal(err)
	}
	if got := lastShape(); !slices.Equal(got, whole(testArrange.Size)) {
		t.Errorf("no map: cut to %+v", got)
	}
	if err := app.OpenPanel(openAtAbout); err != nil {
		t.Fatal(err)
	}
	if got := lastShape(); !slices.Equal(got, whole(lastPlaced(t, seen).Size)) {
		t.Errorf("panel: cut to %+v", got)
	}
	app.shape = func([]placement.Rect) error { return errPlanted }
	placedBefore := len(seen.placed)
	app.place = func(at placement.Point, size placement.Size) error {
		seen.placed = append(seen.placed, arranger.Arrangement{At: at, Size: size})
		return nil
	}
	if err := app.ClosePanel(); err != nil || len(seen.placed) == placedBefore {
		t.Errorf("a failed cut stopped the placing (%v)", err)
	}
	if !strings.Contains(log.String(), "cutting the window") {
		t.Errorf("the failed cut was not logged: %q", log.String())
	}
}

// FR-913: the tab is never cut; it keeps all of itself.
func TestTheTabIsNeverCut(t *testing.T) {
	t.Parallel()
	app, service, seen := unpinnedApp(t)
	service.arrangement = withMap
	if err := app.placeLaunched(); err != nil {
		t.Fatal(err)
	}
	if got := seen.shapes[len(seen.shapes)-1]; !app.Snapshot().Collapsed || !slices.Equal(got, []placement.Rect{{Right: lastPlaced(t, seen).Size.Width, Bottom: lastPlaced(t, seen).Size.Height}}) {
		t.Errorf("tab cut to %+v", got)
	}
}

// FR-901, FR-903: the menu item and the handle each flip their choice, then refit the window.
func TestTheSunMapItemAndTheHandleFlipTheirChoices(t *testing.T) {
	t.Parallel()
	app, service, seen, _ := newTestApp(t)
	app.act(application.ActionSunMap)
	seen.events = nil
	if err := app.TogglePullOut(); err != nil {
		t.Fatal(err)
	}
	if !seen.sawEvent(eventRefresh) {
		t.Error("the handle changed the window but the page was not told to draw it again")
	}
	if !service.settings.SunMap || !service.settings.PullOut {
		t.Errorf("sun map %v, pull out %v; want both on", service.settings.SunMap, service.settings.PullOut)
	}
	if err := app.TogglePullOut(); err != nil || service.settings.PullOut {
		t.Errorf("the second click left the pull out %v (%v)", service.settings.PullOut, err)
	}
}
