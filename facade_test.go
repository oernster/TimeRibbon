package main

import (
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/oernster/timeribbon/internal/application"
	"github.com/oernster/timeribbon/internal/product"
)

// Every change the page can make is followed by fitting the ribbon, whether or not it saved: a
// change whose save failed raises a notice, which is one more cell to fit (FR-707).
func TestEveryChangeFitsTheRibbonAndAnswersTheServicesError(t *testing.T) {
	changes := map[string]func(app *App) error{
		"AddClock":       func(app *App) error { _, err := app.AddClock("Europe/London"); return err },
		"RenameClock":    func(app *App) error { return app.RenameClock("id-1", "Home") },
		"RezoneClock":    func(app *App) error { return app.RezoneClock("id-1", "Asia/Kolkata") },
		"RemoveClock":    func(app *App) error { return app.RemoveClock("id-1") },
		"SetStyle":       func(app *App) error { return app.SetStyle("analogue") },
		"SetSize":        func(app *App) error { return app.SetSize("small") },
		"SetFormat":      func(app *App) error { return app.SetFormat("12h") },
		"SetTheme":       func(app *App) error { return app.SetTheme("dark") },
		"SetAlwaysOnTop": func(app *App) error { return app.SetAlwaysOnTop(true) },
		"SetScrollbar":   func(app *App) error { return app.SetScrollbar(12) },
		"DismissNotices": func(app *App) error { app.DismissNotices(); return nil },
	}
	for name, change := range changes {
		for _, failure := range []error{nil, errPlanted} {
			app, service, seen, _ := newTestApp(t)
			service.changeErr = failure
			if err := change(app); name != "DismissNotices" && !errors.Is(err, failure) {
				t.Errorf("%s answered %v, want the service's %v", name, err, failure)
			}
			if !slices.Contains(service.calls, name) {
				t.Errorf("%s never reached the service", name)
			}
			if !slices.Contains(service.calls, "Rearrange") || len(seen.placed) != 1 {
				t.Errorf("%s (service answered %v) placed the ribbon %d times, want it fitted once", name, failure, len(seen.placed))
			}
		}
	}
}

// While a panel is open the window is that panel, so a change does not fit the ribbon; nor does one
// made before startup has found it.
func TestAChangeLeavesThePanelOrAnUnfoundRibbonAlone(t *testing.T) {
	app, service, seen, _ := newTestApp(t)
	app.panelOpen.Store(true)
	_ = app.RenameClock("id-1", "Home")
	app.panelOpen.Store(false)
	app.ribbon = 0
	_ = app.RenameClock("id-1", "Home")
	if slices.Contains(service.calls, "Rearrange") || len(seen.placed) != 0 {
		t.Errorf("the ribbon was fitted %d times, want none", len(seen.placed))
	}
}

func TestFittingPlacesTheArrangementAndKeepsWhetherItScrolls(t *testing.T) {
	app, service, seen, _ := newTestApp(t)
	_ = app.SetScrollbar(12)
	if service.at[0] != testRibbonAt {
		t.Errorf("fitted from %v, want where the ribbon stands, %v", service.at[0], testRibbonAt)
	}
	if seen.placed[0].At != testArrange.At || seen.placed[0].Size != testArrange.Size {
		t.Errorf("placed %+v, want the service's arrangement %+v", seen.placed[0], testArrange)
	}
	if !app.scrolls.Load() || !app.Snapshot().Scrolls {
		t.Error("the arrangement scrolls but the facade does not say so")
	}
}

// The snapshot reaches the page whole, with an empty list of notices rather than none, since the
// page counts them.
func TestTheSnapshotCarriesEveryCellAndNeverANullNoticeList(t *testing.T) {
	app, service, _, _ := newTestApp(t)
	service.snapshot = application.Snapshot{Cells: []application.Cell{{ID: "id-1", Label: "London", Time: "20:37"}}}
	got := app.Snapshot()
	if len(got.Cells) != 1 || got.Cells[0].ID != "id-1" || got.Cells[0].Time != "20:37" {
		t.Errorf("cells %+v, want the one cell the service answered", got.Cells)
	}
	if got.Notices == nil {
		t.Error("the notices went out as null")
	}
}

func TestAFitThatCannotBeWorkedOutIsLoggedAndNothingMoves(t *testing.T) {
	for _, plant := range []func(*scriptedService, *window){
		func(_ *scriptedService, seen *window) { seen.readErr = errPlanted },
		func(service *scriptedService, _ *window) { service.arrangeErr = errPlanted },
	} {
		app, service, seen, log := newTestApp(t)
		plant(service, seen)
		_ = app.SetScrollbar(12)
		if len(seen.placed) != 0 || log.Len() == 0 {
			t.Errorf("placed %d times with log %q, want nothing placed and the failure logged", len(seen.placed), log)
		}
	}
}

func TestSetAlwaysOnTopAppliesTheSettingAtOnce(t *testing.T) {
	app, _, seen, _ := newTestApp(t)
	_ = app.SetAlwaysOnTop(true)
	if !slices.Equal(seen.onTop, []bool{true}) {
		t.Errorf("the window was set on top %v, want [true]", seen.onTop)
	}
}

func TestOpenPanelCentresThePanelOnTheRibbonsDisplay(t *testing.T) {
	app, service, seen, _ := newTestApp(t)
	if err := app.OpenPanel(); err != nil {
		t.Fatal(err)
	}
	if !app.panelOpen.Load() || service.at[0] != testRibbonAt || service.centred != testPanel {
		t.Errorf("open %v, centred %v from %v; want open, the panel's size from the ribbon", app.panelOpen.Load(), service.centred, service.at)
	}
	if len(seen.placed) != 1 || seen.placed[0].At != testArrange.At {
		t.Errorf("placed %+v, want the centred arrangement", seen.placed)
	}
}

func TestOpenPanelAnswersWhatStoppedIt(t *testing.T) {
	for _, plant := range []func(*scriptedService, *window){
		func(_ *scriptedService, seen *window) { seen.readErr = errPlanted },
		func(service *scriptedService, _ *window) { service.arrangeErr = errPlanted },
		func(_ *scriptedService, seen *window) { seen.placeErr = errPlanted },
	} {
		app, service, seen, _ := newTestApp(t)
		plant(service, seen)
		if err := app.OpenPanel(); !errors.Is(err, errPlanted) {
			t.Errorf("OpenPanel answered %v, want the failure", err)
		}
		if !app.panelOpen.Load() {
			t.Error("a panel that failed to place is no longer counted open, so the ribbon would be fitted over it")
		}
	}
}

func TestClosePanelPutsTheRibbonWhereItWasLastLeft(t *testing.T) {
	app, service, seen, _ := newTestApp(t)
	app.panelOpen.Store(true)
	if err := app.ClosePanel(); err != nil {
		t.Fatal(err)
	}
	if app.panelOpen.Load() || !slices.Contains(service.calls, "Launch") || len(seen.placed) != 1 {
		t.Errorf("open %v after %v, placed %d; want closed, launched and placed", app.panelOpen.Load(), service.calls, len(seen.placed))
	}
	service.arrangeErr = errPlanted
	if err := app.ClosePanel(); !errors.Is(err, errPlanted) {
		t.Errorf("ClosePanel answered %v, want the service's failure", err)
	}
}

func TestShowContextMenuShowsTheServicesMenu(t *testing.T) {
	app, service, seen, _ := newTestApp(t)
	service.menu = []application.MenuItem{{Action: application.ActionHide, Label: "Hide ribbon"}}
	app.ShowContextMenu()
	if len(seen.menus) != 1 || !reflect.DeepEqual(seen.menus[0], service.menu) {
		t.Errorf("showed %v, want %v", seen.menus, service.menu)
	}
}

func TestOpenDonationHandsTheAddressToTheBrowser(t *testing.T) {
	app, _, seen, _ := newTestApp(t)
	if err := app.OpenDonation(); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(seen.browsed, []string{product.DonateURL}) {
		t.Errorf("browsed %v, want the donation address once", seen.browsed)
	}
}

// A browser Windows could not open is reported with the reason and the address, so the page can
// still be reached by hand rather than the button doing nothing.
func TestADonationPageThatCouldNotBeOpenedIsReportedWithItsAddress(t *testing.T) {
	app, _, seen, _ := newTestApp(t)
	seen.browseErr = errPlanted
	err := app.OpenDonation()
	if !errors.Is(err, errPlanted) {
		t.Fatalf("OpenDonation answered %v, want the desktop's refusal", err)
	}
	if !strings.Contains(err.Error(), product.DonateURL) {
		t.Errorf("the refusal %q does not give the address", err)
	}
}

func TestTheReadingsPassThrough(t *testing.T) {
	app, service, _, _ := newTestApp(t)
	service.places = []application.Place{{Zone: "Europe/London", Label: "London", Country: "Britain (UK)"}}
	if got := app.SearchPlaces("lon"); len(got) != 1 || got[0].Zone != "Europe/London" {
		t.Errorf("SearchPlaces answered %+v", got)
	}
	if on, err := app.StartWithWindows(); !on || err != nil {
		t.Errorf("StartWithWindows answered %v, %v", on, err)
	}
	if err := app.SetStartWithWindows(true); err != nil || !slices.Contains(service.calls, "SetStartWithWindows") {
		t.Errorf("SetStartWithWindows answered %v after %v", err, service.calls)
	}
	if about := app.About(); about.Name != product.Name || len(about.Credits) != len(product.Credits) {
		t.Errorf("About answered %+v", about)
	}
	if app.Licence() != licenceText || licenceText == "" {
		t.Error("Licence does not answer the embedded LICENSE")
	}
	app.Hide()
	if app.visible.Load() {
		t.Error("Hide left the ribbon counted visible")
	}
}
