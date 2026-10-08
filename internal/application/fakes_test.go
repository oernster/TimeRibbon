package application

import (
	"errors"
	"fmt"
	"testing"
	"time"
	_ "time/tzdata"

	"github.com/oernster/ribbonkit/domain/placement"
	"github.com/oernster/timeribbon/internal/domain/settings"
	"github.com/oernster/timeribbon/internal/domain/sun"
)

// errPlanted is the failure a fake answers when a test asks it to fail.
var errPlanted = errors.New("planted failure")

type fakeStore struct {
	loaded  Loaded
	loadErr error
	saveErr error
	saved   []settings.Settings
}

func (f *fakeStore) Load() (Loaded, error) { return f.loaded, f.loadErr }

func (f *fakeStore) Save(s settings.Settings) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.saved = append(f.saved, s)
	return nil
}

// last answers the settings saved most recently, failing the test when none were.
func (f *fakeStore) last(t *testing.T) settings.Settings {
	t.Helper()
	if len(f.saved) == 0 {
		t.Fatal("nothing was saved")
	}
	return f.saved[len(f.saved)-1]
}

// realZones resolves through the embedded tz database, as infrastructure does, over a small
// catalogue.
type realZones struct{ places []Place }

func (z realZones) Resolve(zone string) (*time.Location, error) { return time.LoadLocation(zone) }

func (z realZones) Catalogue() []Place { return z.places }

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

type countingIDs struct{ next int }

func (c *countingIDs) NewID() string {
	c.next++
	return fmt.Sprintf("id-%d", c.next)
}

type fakeMonitors struct {
	monitors []placement.Monitor
	err      error
}

func (f fakeMonitors) Monitors() ([]placement.Monitor, error) { return f.monitors, f.err }

type fakeStartup struct {
	enabled bool
	err     error
}

func (f *fakeStartup) Enabled() (bool, error) { return f.enabled, f.err }

func (f *fakeStartup) Enable() error {
	f.enabled = f.err == nil
	return f.err
}

func (f *fakeStartup) Disable() error {
	f.enabled = false
	return f.err
}

// testLayout is the cell geometry the tests arrange with, in DIP.
var testLayout = Layout{
	Digital:    placement.Size{Width: 160, Height: 90},
	Analogue:   placement.Size{Width: 160, Height: 150},
	Prompt:     placement.Size{Width: 160, Height: 190},
	Padding:    8,
	HandleLane: 12,
}

// testLayouts is testLayout for the large size and a smaller one for the small (FR-610).
var testLayouts = Layouts{
	Large: testLayout,
	Small: Layout{
		Digital:    placement.Size{Width: 120, Height: 60},
		Analogue:   placement.Size{Width: 120, Height: 100},
		Prompt:     testLayout.Prompt,
		Padding:    testLayout.Padding,
		HandleLane: testLayout.HandleLane,
	},
}

// Two monitors side by side: a primary at 100 percent and a secondary at 150 percent to its right.
var (
	primaryMonitor = placement.Monitor{
		Device: `\\.\DISPLAY1`, Work: placement.Rect{Right: 1920, Bottom: 1032},
		DPI: placement.BaseDPI, Primary: true,
	}
	secondaryMonitor = placement.Monitor{
		Device: `\\.\DISPLAY2`, Work: placement.Rect{Left: 1920, Right: 4480, Bottom: 1392}, DPI: 144,
	}
)

// rig is a service over fakes, with the fakes kept to inspect.
type rig struct {
	service *Service
	store   *fakeStore
	startup *fakeStartup
}

// rigPlaces is the catalogue a rig searches and marks from, each city where the real one puts it.
var rigPlaces = []Place{
	{Zone: "America/New_York", Label: "New York", Country: "United States", At: sun.Point{Latitude: 40.7142, Longitude: -74.0064}},
	{Zone: "Asia/Kolkata", Label: "Kolkata", Country: "India", At: sun.Point{Latitude: 22.5333, Longitude: 88.3667}},
	{Zone: "Europe/London", Label: "London", Country: "Britain (UK)", At: londonAt},
	{Zone: "America/Indiana/Indianapolis", Label: "Indianapolis", Country: "United States", At: sun.Point{Latitude: 39.7683, Longitude: -86.1581}},
}

// newRig answers a service at 2026-09-27T20:37:00Z over both monitors, loaded from initial.
func newRig(t *testing.T, initial settings.Settings) rig {
	t.Helper()
	return newRigOver(t, initial, rigPlaces)
}

// newRigOver is newRig with places as its catalogue.
func newRigOver(t *testing.T, initial settings.Settings, places []Place) rig {
	t.Helper()
	now, err := time.Parse(time.RFC3339, "2026-09-27T20:37:00Z")
	if err != nil {
		t.Fatal(err)
	}
	store := &fakeStore{loaded: Loaded{Settings: initial}}
	startup := &fakeStartup{}
	service := New(Ports{
		Store:    store,
		Zones:    realZones{places: places},
		Clock:    fixedClock{now: now},
		IDs:      &countingIDs{},
		Monitors: fakeMonitors{monitors: []placement.Monitor{primaryMonitor, secondaryMonitor}},
		Startup:  startup,
	}, testLayouts)
	if err := service.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	return rig{service: service, store: store, startup: startup}
}

// withEntries answers the defaults holding entries in order.
func withEntries(entries ...settings.Entry) settings.Settings {
	s := settings.Defaults()
	for _, entry := range entries {
		s = s.WithClockAdded(entry)
	}
	return s
}
