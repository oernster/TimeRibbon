package main

import (
	"errors"
	"slices"
	"testing"

	"github.com/oernster/timeribbon/ribbonkit/application/menus"
	"github.com/oernster/timeribbon/ribbonkit/domain/ribbon"
)

// testChoices is a Colour group of two schemes, Neon ticked, then Pin ribbon on its own.
var testChoices = []menus.Item{
	{Label: "Colour", Children: []menus.Item{
		{Action: "colour-classic", Label: "Classic", Checkable: true},
		{Action: "colour-neon", Label: "Neon", Checkable: true, Checked: true},
	}},
	{Action: menus.Pin, Label: "Pin ribbon", Checkable: true, Checked: true},
}

// FR-624: the snapshot carries the menus' choices whole with their ticks; never a null list.
func TestTheSnapshotCarriesTheMenusChoices(t *testing.T) {
	app, service, _, _ := newTestApp(t)
	service.choices = testChoices
	got := app.Snapshot().Choices
	if len(got) != len(testChoices) || got[0].Label != "Colour" || len(got[0].Children) != 2 {
		t.Fatalf("choices %+v, want the service's", got)
	}
	if neon := got[0].Children[1]; neon.Action != "colour-neon" || !neon.Checkable || !neon.Checked {
		t.Errorf("Neon went out as %+v", neon)
	}
	if pin := got[1]; pin.Children == nil || pin.Action != string(menus.Pin) || !pin.Checked {
		t.Errorf("Pin ribbon went out as %+v, with a null list of children or without its tick", pin)
	}
}

// FR-624: a choice made in Settings is carried out as its menu item is; anything else is refused and
// changes nothing, a group's own label included, since it chooses nothing.
func TestChooseCarriesOutOnlyTheChoicesSettingsOffers(t *testing.T) {
	app, service, seen, _ := newTestApp(t)
	service.choices = testChoices
	if err := app.Choose("colour-classic"); err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(service.calls, "SetColour") || !seen.sawEvent(eventRefresh) {
		t.Errorf("calls %v; want the colour set and the page told", service.calls)
	}
	for _, refused := range []string{"", "exit", "colour-sunset"} {
		app, service, _, _ := newTestApp(t)
		service.choices = testChoices
		if err := app.Choose(refused); !errors.Is(err, ribbon.ErrUnknownChoice) || len(service.calls) != 0 {
			t.Errorf("Choose(%q) answered %v with calls %v; want it refused and nothing done", refused, err, service.calls)
		}
	}
}

// FR-625: Settings opens at its own width and keeps it as its height is fitted; the other panels at
// theirs. A word naming no panel is refused and opens nothing.
func TestSettingsOpensAndFitsAtItsOwnWidth(t *testing.T) {
	app, service, _, _ := newTestApp(t)
	if err := app.OpenPanel(openAtSettings); err != nil || service.centred != testSettingsPanel {
		t.Fatalf("Settings centred at %v (%v), want %v", service.centred, err, testSettingsPanel)
	}
	if err := app.FitPanel(fittedHeight); err != nil || service.centred.Width != testSettingsPanel.Width {
		t.Errorf("fitted at %v (%v), want Settings' width kept", service.centred, err)
	}
	for _, other := range []string{openAtAbout, openAtLicence, openAtUpdate} {
		app, service, _, _ := newTestApp(t)
		if err := app.OpenPanel(other); err != nil || service.centred != testPanel {
			t.Errorf("%s centred at %v (%v), want %v", other, service.centred, err, testPanel)
		}
	}
	app, service, seen, _ := newTestApp(t)
	if err := app.OpenPanel("sideboard"); !errors.Is(err, ribbon.ErrUnknownChoice) || app.panelOpen.Load() || len(seen.placed) != 0 || len(service.at) != 0 {
		t.Errorf("an unknown panel answered %v, open %v, placed %d times", err, app.panelOpen.Load(), len(seen.placed))
	}
}
