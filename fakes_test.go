package main

import (
	"errors"
	"testing"

	"github.com/oernster/ribbonkit/application/menus"
	"github.com/oernster/ribbonkit/ui/window"
	"github.com/oernster/timeribbon/internal/application"
	"github.com/oernster/timeribbon/internal/domain/clock"
	"github.com/oernster/timeribbon/internal/domain/settings"
)

// errPlanted is the failure a stand-in answers when a test asks it to fail.
var errPlanted = errors.New("planted failure")

// scriptedService stands in for application.Service in TimeRibbon's own half of the facade. It
// answers what a test sets and records the calls the facade made.
type scriptedService struct {
	settings settings.Settings
	snapshot application.Snapshot
	places   []application.Place
	choices  []menus.Item
	// measured is the last measurement SetMeasured was handed.
	measured application.Measured
	// changeErr answers every change.
	changeErr error
	calls     []string
}

func (s *scriptedService) record(call string) { s.calls = append(s.calls, call) }

func (s *scriptedService) change(call string) error {
	s.record(call)
	return s.changeErr
}

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

func (s *scriptedService) SetFormat(clock.Format) error { return s.change("SetFormat") }

func (s *scriptedService) SetDateFormat(clock.DateFormat) error { return s.change("SetDateFormat") }

func (s *scriptedService) SetSunMap(on bool) error {
	err := s.change("SetSunMap")
	if err == nil {
		s.settings.SunMap = on
	}
	return err
}

func (s *scriptedService) DismissNotices() { s.record("DismissNotices") }

func (s *scriptedService) TextSamples() (times, dates []string) {
	return []string{"23:59"}, []string{"Wednesday, 30 September"}
}

func (s *scriptedService) SetMeasured(measured application.Measured) error {
	s.measured = measured
	return s.change("SetMeasured")
}

func (s *scriptedService) SettingsChoices() []menus.Item { return s.choices }

// recordingControl stands in for the window, recording what TimeRibbon's half asked of it in order.
type recordingControl struct {
	calls []string
	// reported is each failure reported, as its doing; panels each panel shown.
	reported []string
	panels   []string
	colours  []string
	turned   []string
	shown    window.Shown
	// choiceErr answers SetColour and SetOrientation.
	choiceErr error
}

func (c *recordingControl) record(call string) { c.calls = append(c.calls, call) }

func (c *recordingControl) Refitted(err error) error {
	c.record("Refitted")
	return err
}

func (c *recordingControl) ContentChanged() { c.record("ContentChanged") }

func (c *recordingControl) Redraw() { c.record("Redraw") }

func (c *recordingControl) Redrawn(err error) error {
	c.record("Redrawn")
	return err
}

func (c *recordingControl) Report(doing string, err error) {
	if err != nil {
		c.reported = append(c.reported, doing)
	}
}

func (c *recordingControl) ShowPanel(panel string) { c.panels = append(c.panels, panel) }

func (c *recordingControl) PageMeasured() { c.record("PageMeasured") }

func (c *recordingControl) Shown() window.Shown { return c.shown }

func (c *recordingControl) SetColour(colour string) error {
	c.colours = append(c.colours, colour)
	return c.choiceErr
}

func (c *recordingControl) SetOrientation(orientation string) error {
	c.turned = append(c.turned, orientation)
	return c.choiceErr
}

// fitted counts how often the facade had the window fit the ribbon.
func (c *recordingControl) fitted() int {
	count := 0
	for _, call := range c.calls {
		if call == "Refitted" || call == "ContentChanged" {
			count++
		}
	}
	return count
}

// newTestApp answers TimeRibbon's half of the facade over a scripted service, its window the
// recording stand-in answered with it.
func newTestApp(t *testing.T) (*App, *scriptedService, *recordingControl) {
	t.Helper()
	service, control := &scriptedService{}, &recordingControl{}
	return &App{service: service, control: control}, service, control
}
