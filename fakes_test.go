package main

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/oernster/timestrip/internal/application"
	"github.com/oernster/timestrip/internal/domain/clock"
	"github.com/oernster/timestrip/internal/domain/placement"
	"github.com/oernster/timestrip/internal/domain/settings"
)

// errPlanted is the failure a stand-in answers when a test asks it to fail.
var errPlanted = errors.New("planted failure")

// scriptedService stands in for application.Service. It answers what a test sets and records the
// calls the facade made, so a test reads what the facade decided with each answer.
type scriptedService struct {
	settings    settings.Settings
	snapshot    application.Snapshot
	places      []application.Place
	menu        []application.MenuItem
	arrangement application.Arrangement
	// changeErr answers every change; arrangeErr and movedErr answer the arranging calls.
	changeErr  error
	arrangeErr error
	movedErr   error
	// panicOnRearrange makes Rearrange panic, for the facade's recover.
	panicOnRearrange bool

	calls   []string
	at      []placement.Point
	centred placement.Size
	onTop   []bool
}

func (s *scriptedService) record(call string) { s.calls = append(s.calls, call) }

func (s *scriptedService) Snapshot() application.Snapshot { return s.snapshot }

func (s *scriptedService) Settings() settings.Settings { return s.settings }

func (s *scriptedService) AddClock(string) (string, error) {
	s.record("AddClock")
	return "id-1", s.changeErr
}

func (s *scriptedService) RenameClock(string, string) error { return s.change("RenameClock") }

func (s *scriptedService) RezoneClock(string, string) error { return s.change("RezoneClock") }

func (s *scriptedService) RemoveClock(string) error { return s.change("RemoveClock") }

func (s *scriptedService) MoveClock(string, int) error { return s.change("MoveClock") }

func (s *scriptedService) SearchPlaces(string) []application.Place { return s.places }

func (s *scriptedService) SetStyle(settings.Style) error { return s.change("SetStyle") }

func (s *scriptedService) SetFormat(clock.Format) error { return s.change("SetFormat") }

func (s *scriptedService) SetOrientation(settings.Orientation) error {
	return s.change("SetOrientation")
}

func (s *scriptedService) SetTheme(settings.Theme) error { return s.change("SetTheme") }

func (s *scriptedService) SetAlwaysOnTop(on bool) error {
	s.onTop = append(s.onTop, on)
	s.settings.AlwaysOnTop = on
	return s.change("SetAlwaysOnTop")
}

func (s *scriptedService) StartWithWindows() (bool, error) { return true, s.changeErr }

func (s *scriptedService) SetStartWithWindows(bool) error { return s.change("SetStartWithWindows") }

func (s *scriptedService) DismissNotices() { s.record("DismissNotices") }

func (s *scriptedService) SetScrollbar(int) error { return s.change("SetScrollbar") }

func (s *scriptedService) ContextMenu() []application.MenuItem { return s.menu }

func (s *scriptedService) CloseRequested() application.MenuAction { return application.ActionHide }

func (s *scriptedService) Launch() (application.Arrangement, error) {
	s.record("Launch")
	return s.arrangement, s.arrangeErr
}

func (s *scriptedService) Rearrange(at placement.Point) (application.Arrangement, error) {
	if s.panicOnRearrange {
		panic("planted panic")
	}
	s.record("Rearrange")
	s.at = append(s.at, at)
	return s.arrangement, s.arrangeErr
}

func (s *scriptedService) Moved(at placement.Point) (application.Arrangement, error) {
	s.record("Moved")
	s.at = append(s.at, at)
	return s.arrangement, s.movedErr
}

func (s *scriptedService) Centred(at placement.Point, size placement.Size) (application.Arrangement, error) {
	s.record("Centred")
	s.at = append(s.at, at)
	s.centred = size
	return s.arrangement, s.arrangeErr
}

func (s *scriptedService) change(call string) error {
	s.record(call)
	return s.changeErr
}

// emitted is one event the facade sent the page.
type emitted struct {
	event string
	data  []any
}

// window records what the facade asked of Wails and the desktop, standing in for both.
type window struct {
	events    []emitted
	shown     int
	hidden    int
	quits     int
	onTop     []bool
	browsed   []string
	menus     [][]application.MenuItem
	placed    []application.Arrangement
	stripAt   placement.Point
	readErr   error
	placeErr  error
	browseErr error
	positions int
}

// sawEvent reports whether the facade sent event with data first, when data is given.
func (w *window) sawEvent(event string, data ...any) bool {
	for _, sent := range w.events {
		if sent.event == event && (len(data) == 0 || (len(sent.data) > 0 && sent.data[0] == data[0])) {
			return true
		}
	}
	return false
}

// Where the tests' strip stands, the arrangement the stand-in service answers and the panel's size.
var (
	testStripAt = placement.Point{X: 40, Y: 60}
	testArrange = application.Arrangement{
		At: placement.Point{X: 10, Y: 20}, Size: placement.Size{Width: 300, Height: 90}, Scrolls: true,
	}
	testPanel = placement.Size{Width: 560, Height: 760}
)

// testStrip is the strip's window handle once startup has found it: any value that is not none.
const testStrip = 1

// newTestApp answers a facade over a scripted service, started and with its strip found, whose
// Wails and desktop calls land in the window answered with it.
func newTestApp(t *testing.T) (*App, *scriptedService, *window, *bytes.Buffer) {
	t.Helper()
	service := &scriptedService{arrangement: testArrange}
	seen := &window{stripAt: testStripAt}
	log := &bytes.Buffer{}
	app := newApp(service, nil, log, testPanel)
	app.ctx = context.Background()
	app.strip = testStrip
	app.emit = func(event string, data ...any) { seen.events = append(seen.events, emitted{event, data}) }
	app.showWindow = func() { seen.shown++ }
	app.hideWindow = func() { seen.hidden++ }
	app.quit = func() { seen.quits++ }
	app.setOnTop = func(on bool) { seen.onTop = append(seen.onTop, on) }
	app.browse = func(address string) error {
		seen.browsed = append(seen.browsed, address)
		return seen.browseErr
	}
	app.showMenu = func(items []application.MenuItem) { seen.menus = append(seen.menus, items) }
	app.position = func() (placement.Point, error) {
		seen.positions++
		return seen.stripAt, seen.readErr
	}
	app.place = func(at placement.Point, size placement.Size) error {
		seen.placed = append(seen.placed, application.Arrangement{At: at, Size: size})
		return seen.placeErr
	}
	return app, service, seen, log
}
