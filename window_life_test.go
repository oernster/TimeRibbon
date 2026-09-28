package main

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/oernster/timeribbon/internal/application"
	"github.com/oernster/timeribbon/internal/domain/placement"
	"github.com/oernster/timeribbon/internal/infrastructure/desktop"
)

func TestADragIsRecordedAndTheRibbonPlacedWhereTheServiceSays(t *testing.T) {
	app, service, seen, _ := newTestApp(t)
	app.handleSafely(desktop.Event{Kind: desktop.EventMoveEnded})
	if !slices.Equal(service.calls, []string{"Moved"}) || service.at[0] != testRibbonAt {
		t.Errorf("the service heard %v from %v, want Moved from where the drag let go", service.calls, service.at)
	}
	if len(seen.placed) != 1 || seen.placed[0].At != testArrange.At || !app.scrolls.Load() {
		t.Errorf("placed %+v, want the recorded arrangement", seen.placed)
	}
}

// A drag whose placement could not be saved raised a notice: the ribbon is still fitted where it
// stands, its new cell included, rather than left wherever the drag let go (FR-707).
func TestADragWhoseSaveFailedIsStillFitted(t *testing.T) {
	app, service, seen, log := newTestApp(t)
	service.movedErr = errPlanted
	app.moved()
	if !slices.Equal(service.calls, []string{"Moved", "Rearrange"}) || len(seen.placed) != 1 {
		t.Errorf("the service heard %v and the ribbon was placed %d times, want it fitted once", service.calls, len(seen.placed))
	}
	if !strings.Contains(log.String(), "recording where the ribbon was left") {
		t.Errorf("the failed save was not logged: %q", log)
	}
}

func TestADragIsIgnoredWhileAPanelIsOpenOrWhenTheRibbonCannotBeRead(t *testing.T) {
	app, service, seen, log := newTestApp(t)
	app.panelOpen.Store(true)
	app.moved()
	app.panelOpen.Store(false)
	seen.readErr = errPlanted
	app.moved()
	if len(service.calls) != 0 || len(seen.placed) != 0 {
		t.Errorf("the service heard %v and the ribbon was placed %d times, want neither", service.calls, len(seen.placed))
	}
	if !strings.Contains(log.String(), "reading where the ribbon was left") {
		t.Errorf("the unreadable position was not logged: %q", log)
	}
}

func TestTheDesktopsEventsRefitOrRefresh(t *testing.T) {
	app, service, seen, _ := newTestApp(t)
	app.handleSafely(desktop.Event{Kind: desktop.EventDisplayChanged})
	if !slices.Contains(service.calls, "Rearrange") || !seen.sawEvent(eventRefresh) {
		t.Errorf("a display change reached %v and sent %v, want a fit and a refresh", service.calls, seen.events)
	}
	app, service, seen, _ = newTestApp(t)
	app.panelOpen.Store(true)
	app.handleSafely(desktop.Event{Kind: desktop.EventDisplayChanged})
	if len(service.calls) != 0 || len(seen.placed) != 0 {
		t.Errorf("a display change under an open panel reached %v, want the panel left alone", service.calls)
	}
	for _, kind := range []desktop.EventKind{desktop.EventTimeChanged, desktop.EventResumed} {
		app, _, seen, _ := newTestApp(t)
		app.handleSafely(desktop.Event{Kind: kind})
		if !seen.sawEvent(eventRefresh) {
			t.Errorf("desktop event %d sent %v, want a refresh", kind, seen.events)
		}
	}
}

// FR-506, Amendment 17: each further launch flips the ribbon, as the tray icon's click does.
func TestASecondLaunchTogglesTheRibbon(t *testing.T) {
	app, _, seen, _ := newTestApp(t)
	app.show()
	app.secondInstance()
	if seen.hidden != 1 || app.visible.Load() {
		t.Fatalf("hidden %d, visible %v; want a shown ribbon hidden", seen.hidden, app.visible.Load())
	}
	app.secondInstance()
	if seen.shown != 2 || !app.visible.Load() {
		t.Errorf("shown %d, visible %v; want the hidden ribbon shown again", seen.shown, app.visible.Load())
	}
}

func TestTheTrayIconTogglesTheRibbon(t *testing.T) {
	app, _, seen, _ := newTestApp(t)
	app.handleSafely(desktop.Event{Kind: desktop.EventIconClicked})
	app.handleSafely(desktop.Event{Kind: desktop.EventIconClicked})
	if seen.shown != 1 || seen.hidden != 1 || app.visible.Load() {
		t.Errorf("shown %d, hidden %d, visible %v; want shown then hidden", seen.shown, seen.hidden, app.visible.Load())
	}
}

// A panic in one event is logged and the next is still heard (the listen loop's promise).
func TestAPanicInOneEventIsLoggedAndSurvived(t *testing.T) {
	app, service, _, log := newTestApp(t)
	service.panicOnRearrange = true
	app.handleSafely(desktop.Event{Kind: desktop.EventDisplayChanged})
	if !strings.Contains(log.String(), "recovered from planted panic") {
		t.Errorf("the panic was not logged: %q", log)
	}
}

func TestEachMenuActionOpensWhatItNames(t *testing.T) {
	panels := map[application.MenuAction]string{
		application.ActionAddClock: openAtAddClock,
		application.ActionSettings: openAtSettings,
		application.ActionAbout:    openAtAbout,
		application.ActionLicence:  openAtLicence,
	}
	for action, panel := range panels {
		app, _, seen, _ := newTestApp(t)
		app.handleSafely(desktop.Event{Kind: desktop.EventMenu, Action: action})
		if seen.shown != 1 || !seen.sawEvent(eventOpenPanel, panel) {
			t.Errorf("action %v showed %d times and sent %v, want the ribbon shown and %s opened", action, seen.shown, seen.events, panel)
		}
	}
}

func TestShowAndHideFromTheMenu(t *testing.T) {
	app, _, seen, _ := newTestApp(t)
	app.act(application.ActionShow)
	if seen.shown != 1 || !app.visible.Load() || !seen.sawEvent(eventRefresh) {
		t.Errorf("shown %d, visible %v, sent %v; want shown, visible and refreshed", seen.shown, app.visible.Load(), seen.events)
	}
	app.act(application.ActionHide)
	if seen.hidden != 1 || app.visible.Load() {
		t.Errorf("hidden %d, visible %v; want hidden", seen.hidden, app.visible.Load())
	}
}

func TestAlwaysOnTopFromTheMenuTurnsTheSettingOver(t *testing.T) {
	app, service, seen, _ := newTestApp(t)
	service.settings.AlwaysOnTop = true
	app.act(application.ActionAlwaysOnTop)
	if !slices.Equal(service.onTop, []bool{false}) || !slices.Equal(seen.onTop, []bool{false}) {
		t.Errorf("the service was set %v and the window %v, want both turned off", service.onTop, seen.onTop)
	}
	if !seen.sawEvent(eventRefresh) {
		t.Error("the page was not told the setting changed")
	}
}

// FR-408: a Position item puts the ribbon where the service says and shows it; under an open panel
// the place is kept for the panel's close and the panel is left where it is.
func TestAPositionItemPutsTheRibbonAgainstItsEdge(t *testing.T) {
	app, service, seen, _ := newTestApp(t)
	app.act(application.ActionRightEdge)
	if !slices.Equal(service.calls, []string{"ToEdge"}) || service.at[0] != testRibbonAt || service.edges[0] != placement.Right {
		t.Errorf("the service heard %v at %v for %v, want ToEdge from the ribbon for the right edge", service.calls, service.at, service.edges)
	}
	if len(seen.placed) != 1 || seen.placed[0].At != testArrange.At || seen.shown != 1 || !app.scrolls.Load() {
		t.Errorf("placed %+v and shown %d times, want placed where the service said and shown", seen.placed, seen.shown)
	}
	app, service, seen, _ = newTestApp(t)
	app.panelOpen.Store(true)
	app.act(application.ActionTopEdge)
	if !slices.Equal(service.calls, []string{"ToEdge"}) || len(seen.placed) != 0 {
		t.Errorf("under a panel the service heard %v and the window was placed %d times, want the place kept only", service.calls, len(seen.placed))
	}
}

// FR-409: choosing an orientation puts the ribbon against its home edge, even where the save failed,
// since the choice took; a refused choice fits the ribbon where it stands instead.
func TestChoosingAnOrientationGoesToItsHomeEdge(t *testing.T) {
	for orientation, edge := range map[string]placement.Edge{"horizontal": placement.Top, "vertical": placement.Right} {
		for _, failure := range []error{nil, errPlanted} {
			app, service, seen, _ := newTestApp(t)
			service.changeErr = failure
			if err := app.SetOrientation(orientation); !errors.Is(err, failure) {
				t.Errorf("%s answered %v, want %v", orientation, err, failure)
			}
			if !slices.Equal(service.calls, []string{"SetOrientation", "ToEdge"}) || service.edges[0] != edge || len(seen.placed) != 1 {
				t.Errorf("%s (save %v): the service heard %v for %v, placed %d times; want %s once", orientation, failure, service.calls, service.edges, len(seen.placed), edge)
			}
		}
	}
	app, service, seen, _ := newTestApp(t)
	service.changeErr = application.ErrUnknownChoice
	if err := app.SetOrientation("diagonal"); !errors.Is(err, application.ErrUnknownChoice) {
		t.Errorf("a refused orientation answered %v", err)
	}
	if !slices.Equal(service.calls, []string{"SetOrientation", "Rearrange"}) || len(seen.placed) != 1 {
		t.Errorf("a refused orientation reached %v and placed %d times, want it fitted where it stands", service.calls, len(seen.placed))
	}
}

// FR-108: the Style and Orientation items choose through the facade and have the page redraw.
func TestStyleAndOrientationItemsChooseAndRedraw(t *testing.T) {
	app, service, seen, _ := newTestApp(t)
	app.act(application.ActionAnalogue)
	if !slices.Equal(service.calls, []string{"SetStyle", "Rearrange"}) || !seen.sawEvent(eventRefresh) {
		t.Errorf("Analogue reached %v and sent %v, want the style set, the ribbon fitted and a redraw", service.calls, seen.events)
	}
	app, service, seen, _ = newTestApp(t)
	app.act("colour-neon")
	if !slices.Equal(service.calls, []string{"SetColour", "Rearrange"}) || !seen.sawEvent(eventRefresh) {
		t.Errorf("Neon reached %v and sent %v, want the colour set, the ribbon fitted and a redraw", service.calls, seen.events)
	}
	app, service, seen, _ = newTestApp(t)
	app.act(application.ActionHorizontal)
	if !slices.Equal(service.calls, []string{"SetOrientation", "ToEdge"}) || service.edges[0] != placement.Top || !seen.sawEvent(eventRefresh) {
		t.Errorf("Horizontal reached %v for %v and sent %v, want it set and the ribbon at the top", service.calls, service.edges, seen.events)
	}
	app, service, _, _ = newTestApp(t)
	app.act("no-such-action")
	if len(service.calls) != 0 {
		t.Errorf("an unknown action reached %v", service.calls)
	}
}

// Before startup has found the ribbon there is nothing to put against an edge.
func TestNoRibbonIsMovedBeforeStartupFindsIt(t *testing.T) {
	app, service, seen, _ := newTestApp(t)
	app.ribbon = 0
	app.act(application.ActionLeftEdge)
	if len(service.calls) != 0 || len(seen.placed) != 0 {
		t.Errorf("the service heard %v and the window was placed %d times", service.calls, len(seen.placed))
	}
}

// A Position item whose place cannot be read or decided moves nothing and says why in the log.
func TestAPositionItemThatFailsMovesNothing(t *testing.T) {
	app, service, seen, log := newTestApp(t)
	seen.readErr = errPlanted
	app.act(application.ActionLeftEdge)
	service.arrangeErr = errPlanted
	seen.readErr = nil
	app.act(application.ActionLeftEdge)
	if len(seen.placed) != 0 || !strings.Contains(log.String(), "reading where the ribbon is") ||
		!strings.Contains(log.String(), "putting the ribbon against an edge") {
		t.Errorf("placed %d times with log %q, want nothing placed and both failures logged", len(seen.placed), log)
	}
}

func TestExitQuitsAndLetsTheCloseThrough(t *testing.T) {
	app, _, seen, _ := newTestApp(t)
	app.trayUp.Store(true)
	app.act(application.ActionExit)
	if seen.quits != 1 || !app.quitting.Load() {
		t.Errorf("quit %d times, quitting %v; want one quit decided", seen.quits, app.quitting.Load())
	}
	if app.beforeClose(context.Background()) {
		t.Error("the close after Exit was held back, so the application would not end")
	}
}

// A close such as Alt+F4 hides the ribbon while a tray icon can bring it back (FR-507); with none,
// the close goes ahead rather than leave a running program with nothing on screen.
func TestACloseHidesTheRibbonOnlyWhileTheTrayIsUp(t *testing.T) {
	app, _, seen, _ := newTestApp(t)
	if app.beforeClose(context.Background()) {
		t.Error("the close was held back with no tray icon to bring the ribbon back from")
	}
	app.trayUp.Store(true)
	if !app.beforeClose(context.Background()) || seen.hidden != 1 {
		t.Errorf("the close went ahead or hid %d times, want it held and the ribbon hidden", seen.hidden)
	}
}

func TestNothingReachesTheWindowBeforeStartup(t *testing.T) {
	app, _, seen, _ := newTestApp(t)
	app.ctx = nil
	app.secondInstance()
	app.hide()
	app.applyAlwaysOnTop()
	app.act(application.ActionExit)
	if seen.shown+seen.hidden+seen.quits+len(seen.onTop) != 0 {
		t.Errorf("before startup the window was shown %d, hidden %d, quit %d and set on top %v", seen.shown, seen.hidden, seen.quits, seen.onTop)
	}
}

func TestDomReadyShowsTheRibbon(t *testing.T) {
	app, _, seen, _ := newTestApp(t)
	app.domReady(context.Background())
	if seen.shown != 1 {
		t.Errorf("shown %d times, want once the page has drawn", seen.shown)
	}
}
