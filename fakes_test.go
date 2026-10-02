package main

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/oernster/timeribbon/internal/application"
	"github.com/oernster/timeribbon/internal/domain/clock"
	"github.com/oernster/timeribbon/internal/domain/placement"
	"github.com/oernster/timeribbon/internal/domain/settings"
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
	choices     []application.MenuItem
	arrangement application.Arrangement
	// measured is the last measurement SetMeasured was handed.
	measured application.Measured
	// lastEdge, when set, is what ToLastEdge answers.
	lastEdge *application.Arrangement
	// changeErr answers every change; arrangeErr and movedErr answer the arranging calls.
	changeErr  error
	arrangeErr error
	movedErr   error
	// panicOnRearrange makes Rearrange panic, for the facade's recover.
	panicOnRearrange bool
	// update answers every update check; panicOnUpdate makes one panic instead. manualChecks records
	// whether each check was asked for; skipped the version each SkipUpdate kept.
	update        application.UpdateStatus
	panicOnUpdate bool
	manualChecks  []bool
	skipped       []string
	// checked, when set, hears each check as it is made, for a check made on a goroutine.
	checked chan bool

	calls   []string
	at      []placement.Point
	centred placement.Size
	onTop   []bool
	edges   []placement.Edge
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

func (s *scriptedService) SearchPlaces(string) []application.Place { return s.places }

func (s *scriptedService) SetStyle(settings.Style) error { return s.change("SetStyle") }

func (s *scriptedService) SetSize(settings.Size) error { return s.change("SetSize") }

func (s *scriptedService) SetColour(settings.Colour) error { return s.change("SetColour") }

func (s *scriptedService) SetFormat(clock.Format) error { return s.change("SetFormat") }

func (s *scriptedService) SetDateFormat(clock.DateFormat) error { return s.change("SetDateFormat") }

// SetOrientation takes the choice unless it is refused as one the setting does not offer, as the
// service does: a save that fails still leaves the choice in effect.
func (s *scriptedService) SetOrientation(orientation settings.Orientation) error {
	if !errors.Is(s.changeErr, application.ErrUnknownChoice) {
		s.settings.Orientation = orientation
	}
	return s.change("SetOrientation")
}

func (s *scriptedService) SetTheme(settings.Theme) error { return s.change("SetTheme") }

func (s *scriptedService) SetAlwaysOnTop(on bool) error {
	s.onTop = append(s.onTop, on)
	s.settings.AlwaysOnTop = on
	return s.change("SetAlwaysOnTop")
}

func (s *scriptedService) SetPinned(on bool) error {
	err := s.change("SetPinned")
	if err == nil {
		s.settings.Pinned = on
	}
	return err
}

func (s *scriptedService) SetSunMap(on bool) error {
	err := s.change("SetSunMap")
	if err == nil {
		s.settings.SunMap = on
	}
	return err
}

func (s *scriptedService) SetPullOut(open bool) error {
	err := s.change("SetPullOut")
	if err == nil {
		s.settings.PullOut = open
	}
	return err
}

// Collapsed answers the tab as a band 8 wide against the arrangement's right side (FR-614).
func (s *scriptedService) Collapsed(full application.Arrangement) (application.Arrangement, error) {
	s.record("Collapsed")
	return application.Arrangement{
		At:   placement.Point{X: full.At.X + full.Size.Width - placement.TabThickness, Y: full.At.Y},
		Size: placement.Size{Width: placement.TabThickness, Height: full.Size.Height},
	}, s.arrangeErr
}

func (s *scriptedService) StartWithWindows() (bool, error) { return true, s.changeErr }

func (s *scriptedService) SetStartWithWindows(bool) error { return s.change("SetStartWithWindows") }

func (s *scriptedService) DismissNotices() { s.record("DismissNotices") }

func (s *scriptedService) SetScrollbar(int) error { return s.change("SetScrollbar") }

func (s *scriptedService) PreviewScale(int) error { return s.change("PreviewScale") }

func (s *scriptedService) SetScale(int) error { return s.change("SetScale") }

func (s *scriptedService) SetOpacity(percent int) error {
	err := s.change("SetOpacity")
	if err == nil {
		s.settings.Opacity = percent
	}
	return err
}

func (s *scriptedService) TextSamples() (times, dates []string) {
	return []string{"23:59"}, []string{"Wednesday, 30 September"}
}

func (s *scriptedService) SetMeasured(measured application.Measured) error {
	s.measured = measured
	return s.change("SetMeasured")
}

func (s *scriptedService) SetPixelsPerDIP(float64) error { return s.change("SetPixelsPerDIP") }

func (s *scriptedService) ContextMenu() []application.MenuItem { return s.menu }

func (s *scriptedService) SettingsChoices() []application.MenuItem { return s.choices }

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

func (s *scriptedService) ToEdge(at placement.Point, edge placement.Edge) (application.Arrangement, error) {
	s.record("ToEdge")
	s.at = append(s.at, at)
	s.edges = append(s.edges, edge)
	return s.arrangement, s.arrangeErr
}

// ToLastEdge answers the arrangement against the last edge: lastEdge where a test set one, else the
// scripted arrangement.
func (s *scriptedService) ToLastEdge(at placement.Point) (application.Arrangement, error) {
	s.record("ToLastEdge")
	s.at = append(s.at, at)
	if s.lastEdge != nil {
		return *s.lastEdge, s.arrangeErr
	}
	return s.arrangement, s.arrangeErr
}

func (s *scriptedService) Centred(at placement.Point, size placement.Size) (application.Arrangement, error) {
	s.record("Centred")
	s.at = append(s.at, at)
	s.centred = size
	return s.arrangement, s.arrangeErr
}

func (s *scriptedService) CheckForUpdate(_ context.Context, manual bool) application.UpdateStatus {
	if s.panicOnUpdate {
		panic("planted panic")
	}
	if s.checked != nil {
		s.checked <- manual
		return s.update
	}
	s.manualChecks = append(s.manualChecks, manual)
	return s.update
}

func (s *scriptedService) SkipUpdate(version string) error {
	s.skipped = append(s.skipped, version)
	return s.change("SkipUpdate")
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
	shapes    [][]placement.Rect
	ribbonAt  placement.Point
	readErr   error
	placeErr  error
	browseErr error
	positions int
	// The unpinned ribbon's calls: the frames asked for, the pointer watching asked for, the time
	// the tests set and the timers pending, which a test fires by hand: hover's, then the fallback
	// that grows an opening ribbon whose page has not said it has drawn.
	tabFrames   []bool
	watching    []bool
	backgrounds [][4]uint8
	now         time.Time
	pending     func()
	waited      time.Duration
	drawPending func()
	// toolkitScale is the toolkit's window scale the desktop reports; unscaled unless a test sets it.
	toolkitScale int
	// sizePending is the launch's fallback for a page that never sizes the ribbon.
	sizePending func()
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

// Where the tests' ribbon stands, the arrangement the stand-in service answers and the panel's size.
var (
	testRibbonAt = placement.Point{X: 40, Y: 60}
	// testArrange stands flush against its right edge, so an unpinned ribbon arranged there is unpinned
	// in effect (FR-619); testAway is the same ribbon standing against no edge.
	testArrange = application.Arrangement{
		At: placement.Point{X: 10, Y: 20}, Size: placement.Size{Width: 300, Height: 90}, Scrolls: true,
		Edge: placement.Right,
	}
	testAway = application.Arrangement{At: placement.Point{X: 400, Y: 300}, Size: testArrange.Size}
	// testPanel is every panel's size but Settings', which is testSettingsPanel.
	testPanel         = placement.Size{Width: 560, Height: 760}
	testSettingsPanel = placement.Size{Width: 900, Height: 760}
)

// testRibbon is the ribbon's window handle once startup has found it: any value that is not none.
const testRibbon = 1

// testUnscaled is the toolkit scale of a desktop that does not scale windows itself.
const testUnscaled = 1

// newTestApp answers a facade over a scripted service, started and with its ribbon found, whose
// Wails and desktop calls land in the window answered with it.
func newTestApp(t *testing.T) (*App, *scriptedService, *window, *bytes.Buffer) {
	t.Helper()
	service := &scriptedService{arrangement: testArrange}
	seen := &window{ribbonAt: testRibbonAt}
	log := &bytes.Buffer{}
	app := newApp(service, nil, log, panelSizes{settings: testSettingsPanel, other: testPanel})
	app.ctx = context.Background()
	app.ribbon = testRibbon
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
		return seen.ribbonAt, seen.readErr
	}
	app.place = func(at placement.Point, size placement.Size) error {
		seen.placed = append(seen.placed, application.Arrangement{At: at, Size: size})
		return seen.placeErr
	}
	app.shape = func(parts []placement.Rect) error {
		seen.shapes = append(seen.shapes, parts)
		return nil
	}
	seen.now = testNow
	app.now = func() time.Time { return seen.now }
	app.after = func(wait time.Duration, do func()) func() bool {
		if wait == drawWait {
			seen.drawPending = do
			return func() bool { seen.drawPending = nil; return true }
		}
		if wait == sizeWait {
			seen.sizePending = do
			return func() bool { seen.sizePending = nil; return true }
		}
		seen.pending, seen.waited = do, wait
		return func() bool { seen.pending = nil; return true }
	}
	app.background = func(red, green, blue, alpha uint8) {
		seen.backgrounds = append(seen.backgrounds, [4]uint8{red, green, blue, alpha})
	}
	app.tabFrame = func(tab bool) error {
		seen.tabFrames = append(seen.tabFrames, tab)
		return nil
	}
	app.watchPointer = func(on bool) { seen.watching = append(seen.watching, on) }
	seen.toolkitScale = testUnscaled
	app.toolkitScale = func() int { return seen.toolkitScale }
	service.settings.Pinned = true
	return app, service, seen, log
}

// testNow is the tests' present moment.
var testNow = time.Date(2026, 9, 28, 21, 0, 0, 0, time.UTC)
