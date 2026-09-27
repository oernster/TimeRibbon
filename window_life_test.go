package main

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/oernster/timestrip/internal/application"
	"github.com/oernster/timestrip/internal/infrastructure/desktop"
)

func TestADragIsRecordedAndTheStripPlacedWhereTheServiceSays(t *testing.T) {
	app, service, seen, _ := newTestApp(t)
	app.handleSafely(desktop.Event{Kind: desktop.EventMoveEnded})
	if !slices.Equal(service.calls, []string{"Moved"}) || service.at[0] != testStripAt {
		t.Errorf("the service heard %v from %v, want Moved from where the drag let go", service.calls, service.at)
	}
	if len(seen.placed) != 1 || seen.placed[0].At != testArrange.At || !app.scrolls.Load() {
		t.Errorf("placed %+v, want the recorded arrangement", seen.placed)
	}
}

// A drag whose placement could not be saved raised a notice: the strip is still fitted where it
// stands, its new cell included, rather than left wherever the drag let go (FR-707).
func TestADragWhoseSaveFailedIsStillFitted(t *testing.T) {
	app, service, seen, log := newTestApp(t)
	service.movedErr = errPlanted
	app.moved()
	if !slices.Equal(service.calls, []string{"Moved", "Rearrange"}) || len(seen.placed) != 1 {
		t.Errorf("the service heard %v and the strip was placed %d times, want it fitted once", service.calls, len(seen.placed))
	}
	if !strings.Contains(log.String(), "recording where the strip was left") {
		t.Errorf("the failed save was not logged: %q", log)
	}
}

func TestADragIsIgnoredWhileAPanelIsOpenOrWhenTheStripCannotBeRead(t *testing.T) {
	app, service, seen, log := newTestApp(t)
	app.panelOpen.Store(true)
	app.moved()
	app.panelOpen.Store(false)
	seen.readErr = errPlanted
	app.moved()
	if len(service.calls) != 0 || len(seen.placed) != 0 {
		t.Errorf("the service heard %v and the strip was placed %d times, want neither", service.calls, len(seen.placed))
	}
	if !strings.Contains(log.String(), "reading where the strip was left") {
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

func TestTheTrayIconTogglesTheStrip(t *testing.T) {
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
			t.Errorf("action %v showed %d times and sent %v, want the strip shown and %s opened", action, seen.shown, seen.events, panel)
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

// A close such as Alt+F4 hides the strip while a tray icon can bring it back (FR-507); with none,
// the close goes ahead rather than leave a running program with nothing on screen.
func TestACloseHidesTheStripOnlyWhileTheTrayIsUp(t *testing.T) {
	app, _, seen, _ := newTestApp(t)
	if app.beforeClose(context.Background()) {
		t.Error("the close was held back with no tray icon to bring the strip back from")
	}
	app.trayUp.Store(true)
	if !app.beforeClose(context.Background()) || seen.hidden != 1 {
		t.Errorf("the close went ahead or hid %d times, want it held and the strip hidden", seen.hidden)
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

func TestDomReadyShowsTheStrip(t *testing.T) {
	app, _, seen, _ := newTestApp(t)
	app.domReady(context.Background())
	if seen.shown != 1 {
		t.Errorf("shown %d times, want once the page has drawn", seen.shown)
	}
}
