package main

import (
	"errors"
	"slices"
	"testing"

	"github.com/oernster/timeribbon/internal/application"
	"github.com/oernster/timeribbon/internal/domain/settings"
	"github.com/oernster/timeribbon/internal/domain/sun"
	"github.com/oernster/timeribbon/internal/product"
	"github.com/oernster/timeribbon/ribbonkit/application/menus"
	"github.com/oernster/timeribbon/ribbonkit/domain/placement"
	"github.com/oernster/timeribbon/ribbonkit/ui/window"
)

// Every change to the clocks is followed by fitting the ribbon, whether or not it saved: a change
// whose save failed raises a notice, which is one more cell to fit (FR-707).
func TestEveryChangeFitsTheRibbonAndAnswersTheServicesError(t *testing.T) {
	changes := map[string]func(app *App) error{
		"AddClock":       func(app *App) error { _, err := app.AddClock("Europe/London"); return err },
		"RenameClock":    func(app *App) error { return app.RenameClock("id-1", "Home") },
		"RezoneClock":    func(app *App) error { return app.RezoneClock("id-1", "Asia/Kolkata") },
		"RemoveClock":    func(app *App) error { return app.RemoveClock("id-1") },
		"SetStyle":       func(app *App) error { return app.SetStyle("analogue") },
		"SetSize":        func(app *App) error { return app.SetSize("small") },
		"SetFormat":      func(app *App) error { return app.SetFormat("12h") },
		"SetDateFormat":  func(app *App) error { return app.SetDateFormat("dmy") },
		"SetMeasured":    func(app *App) error { return app.SetMeasured(measuredDTO{CellWidth: 180}) },
		"SetSunMap":      func(app *App) error { return app.SetSunMap(true) },
		"DismissNotices": func(app *App) error { app.DismissNotices(); return nil },
	}
	for name, change := range changes {
		for _, failure := range []error{nil, errPlanted} {
			app, service, control := newTestApp(t)
			service.changeErr = failure
			if err := change(app); name != "DismissNotices" && !errors.Is(err, failure) {
				t.Errorf("%s answered %v, want the service's %v", name, err, failure)
			}
			if !slices.Contains(service.calls, name) {
				t.Errorf("%s never reached the service", name)
			}
			if control.fitted() != 1 {
				t.Errorf("%s (service answered %v) fitted the ribbon %d times, want once", name, failure, control.fitted())
			}
		}
	}
}

// Turning the sun map on or off changes the window's size, so the page is told to draw it again.
func TestTheSunMapIsRedrawnOnceFitted(t *testing.T) {
	app, _, control := newTestApp(t)
	_ = app.SetSunMap(true)
	if !slices.Equal(control.calls, []string{"Refitted", "Redrawn"}) {
		t.Errorf("the window was asked %v, want the ribbon fitted then redrawn", control.calls)
	}
}

// The snapshot reaches the page whole, with an empty list of notices rather than none, since the
// page counts them. It carries how the window shows the ribbon (FR-614, FR-910).
func TestTheSnapshotCarriesEveryCellAndTheWindowsReading(t *testing.T) {
	app, service, control := newTestApp(t)
	service.snapshot = application.Snapshot{
		Cells:  []application.Cell{{ID: "id-1", Label: "London", Time: "20:37"}},
		SunMap: application.SunMap{On: true, Marks: []application.Mark{{Label: "London", At: sun.Point{Latitude: 51.5, Longitude: -0.1}}}},
	}
	service.choices = []menus.Item{{Action: menus.Pin, Label: "Pin ribbon", Checkable: true}}
	control.shown = window.Shown{
		Collapsed: true, Scrolls: true, DragThreshold: placement.Size{Width: 4, Height: 4},
		PullOutSide: placement.Left, PullOutShown: true,
		Ribbon: window.Box{X: 1, Y: 2, Width: 3, Height: 4}, PullOut: window.Box{Width: 480, Height: 240},
	}
	got := app.Snapshot()
	if len(got.Cells) != 1 || got.Cells[0].ID != "id-1" || got.Cells[0].Time != "20:37" {
		t.Errorf("cells %+v, want the one cell the service answered", got.Cells)
	}
	if got.Notices == nil {
		t.Error("the notices went out as null")
	}
	if !got.Collapsed || !got.Scrolls || got.DragThreshold != (sizeDTO{Width: 4, Height: 4}) {
		t.Errorf("collapsed %v, scrolls %v, threshold %v; want the window's", got.Collapsed, got.Scrolls, got.DragThreshold)
	}
	wantRibbon := window.Box{X: 1, Y: 2, Width: 3, Height: 4}
	if got.SunMap.Side != "left" || !got.SunMap.Shown || got.SunMap.Ribbon != wantRibbon || got.SunMap.Map.Width != 480 {
		t.Errorf("sun map %+v, want the window's layout", got.SunMap)
	}
	if len(got.Choices) != 1 || got.Choices[0].Action != string(menus.Pin) || got.Choices[0].Children == nil {
		t.Errorf("choices %+v, want the service's with never a null list of children", got.Choices)
	}
	if want := (markDTO{Label: "London", Latitude: 51.5, Longitude: -0.1}); !got.SunMap.On || len(got.SunMap.Marks) != 1 || got.SunMap.Marks[0] != want {
		t.Errorf("sun map %v with marks %+v, want it on with London's", got.SunMap.On, got.SunMap.Marks)
	}
}

// The window is handed TimeRibbon's service with the ribbon's own choices and the pull out read out
// of its settings: on a first run, the first-run ones.
func TestTheWindowReadsTheRibbonsChoicesOutOfTheSettings(t *testing.T) {
	kit := kitService{application.New(application.Ports{}, application.Layouts{})}
	first := settings.Defaults()
	if kit.Choices() != first.Choices || kit.PullOut() != first.PullOut {
		t.Errorf("choices %+v, pull out %v; want the first run's %+v, %v", kit.Choices(), kit.PullOut(), first.Choices, first.PullOut)
	}
}

func TestSearchPlacesPassesThrough(t *testing.T) {
	app, service, _ := newTestApp(t)
	service.places = []application.Place{{Zone: "Europe/London", Label: "London", Country: "Britain (UK)"}}
	if got := app.SearchPlaces("lon"); len(got) != 1 || got[0].Zone != "Europe/London" {
		t.Errorf("SearchPlaces answered %+v", got)
	}
}

// TimeRibbon's own menu actions (FR-108, FR-301, FR-901): Add clock opens Settings on the place
// search; every choice is made, a failure reported and the page told to redraw.
func TestTimeRibbonsMenuActions(t *testing.T) {
	app, service, control := newTestApp(t)
	app.actOn(application.ActionAddClock)
	if !slices.Equal(control.panels, []string{openAtAddClock}) {
		t.Errorf("Add clock opened %v, want the place search", control.panels)
	}
	app.actOn(application.ActionSunMap)
	if !service.settings.SunMap {
		t.Error("the Sun map item left the sun map off")
	}
	app.actOn(application.ActionAnalogue)
	app.actOn(menus.Action("colour-neon"))
	app.actOn(application.ActionHorizontal)
	if !slices.Contains(service.calls, "SetStyle") || !slices.Equal(control.colours, []string{"neon"}) || !slices.Equal(control.turned, []string{"horizontal"}) {
		t.Errorf("style %v, colours %v, orientations %v; want each choice made", service.calls, control.colours, control.turned)
	}
	redraws := 0
	for _, call := range control.calls {
		if call == "Redraw" {
			redraws++
		}
	}
	if redraws != 4 {
		t.Errorf("the page was told to redraw %d times, want once for each of the four choices", redraws)
	}
}

func TestAMenuChoiceThatFailedIsReported(t *testing.T) {
	app, service, control := newTestApp(t)
	service.changeErr, control.choiceErr = errPlanted, errPlanted
	for _, action := range []menus.Action{application.ActionSunMap, application.ActionDigital, "colour-ocean", application.ActionVertical} {
		app.actOn(action)
	}
	want := []string{"turning the sun map on or off", "changing the style", "changing the colour", "changing the orientation"}
	if !slices.Equal(control.reported, want) {
		t.Errorf("reported %v, want %v", control.reported, want)
	}
}

// An action that is neither the kit's nor TimeRibbon's changes nothing.
func TestAnUnknownActionChangesNothing(t *testing.T) {
	app, service, control := newTestApp(t)
	app.actOn("no-such-action")
	if len(service.calls) != 0 || len(control.calls) != 0 || len(control.panels) != 0 {
		t.Errorf("service %v, window %v, panels %v; want nothing", service.calls, control.calls, control.panels)
	}
}

// What Help shows comes from internal/product and the LICENSE file (FR-607, FR-608).
func TestTheWindowIsHandedTimeRibbonsProduct(t *testing.T) {
	got := productOf()
	if got.App != product.App() || got.WindowClass != product.RibbonClass || got.Version != product.Version {
		t.Errorf("names %+v, %q, %q; want internal/product's", got.App, got.WindowClass, got.Version)
	}
	if len(got.Credits) != len(product.Credits) || got.DonateURL != product.DonateURL {
		t.Errorf("credits %d, donation %q; want internal/product's", len(got.Credits), got.DonateURL)
	}
	if got.Licence != licenceText || licenceText == "" {
		t.Error("the licence is not the embedded LICENSE")
	}
}
