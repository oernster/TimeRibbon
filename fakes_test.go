package main

import (
	"errors"
	"testing"

	"github.com/oernster/ribbonkit/application/menus"
	"github.com/oernster/ribbonkit/domain/localtime"
	"github.com/oernster/ribbonkit/ui/window/windowtest"
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

func (s *scriptedService) SetFormat(localtime.Format) error { return s.change("SetFormat") }

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

// fitted counts how often the facade had the window fit the ribbon.
func fitted(control *windowtest.Control) int {
	return control.Count("Refitted") + control.Count("ContentChanged")
}

// newTestApp answers TimeRibbon's half of the facade over a scripted service, its window the
// recording stand-in answered with it.
func newTestApp(t *testing.T) (*App, *scriptedService, *windowtest.Control) {
	t.Helper()
	service, control := &scriptedService{}, &windowtest.Control{}
	return &App{service: service, control: control}, service, control
}
