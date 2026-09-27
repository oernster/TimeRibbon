package application

import (
	"slices"
	"testing"

	"github.com/oernster/timestrip/internal/domain/settings"
)

func labels(items []MenuItem) []string {
	var out []string
	for _, item := range items {
		out = append(out, item.Label)
	}
	return out
}

// FR-502.
func TestTrayMenuNamesTheOppositeOfTheVisibility(t *testing.T) {
	t.Parallel()
	r := newRig(t, settings.Defaults())
	shown := r.service.TrayMenu(true)
	if !slices.Equal(labels(shown), []string{"Hide strip", "Add clock", "Settings", "Always on top", "Help", "Exit"}) ||
		shown[0].Action != ActionHide {
		t.Errorf("visible: %+v", shown)
	}
	hidden := r.service.TrayMenu(false)
	if hidden[0].Label != "Show strip" || hidden[0].Action != ActionShow {
		t.Errorf("hidden: %+v", hidden[0])
	}
}

// FR-505, FR-502: the Always on top item shows its state.
func TestAlwaysOnTopItemShowsItsState(t *testing.T) {
	t.Parallel()
	r := newRig(t, settings.Defaults())
	item := r.service.TrayMenu(true)[3]
	if !item.Checkable || item.Checked {
		t.Errorf("off: %+v", item)
	}
	if err := r.service.SetAlwaysOnTop(true); err != nil {
		t.Fatal(err)
	}
	if item := r.service.ContextMenu()[2]; !item.Checked || item.Action != ActionAlwaysOnTop {
		t.Errorf("on: %+v", item)
	}
}

// FR-108.
func TestContextMenuOffersTheStripsActions(t *testing.T) {
	t.Parallel()
	r := newRig(t, settings.Defaults())
	if got := labels(r.service.ContextMenu()); !slices.Equal(got, []string{"Add clock", "Settings", "Always on top", "Help", "Hide strip"}) {
		t.Errorf("got %v", got)
	}
}

// FR-508: Help is a submenu with no action of its own, holding About and Licence, in both menus.
func TestBothMenusOfferHelpWithAboutAndLicence(t *testing.T) {
	t.Parallel()
	r := newRig(t, settings.Defaults())
	for name, menu := range map[string][]MenuItem{"tray": r.service.TrayMenu(true), "context": r.service.ContextMenu()} {
		index := slices.IndexFunc(menu, func(item MenuItem) bool { return item.Label == "Help" })
		if index < 0 {
			t.Fatalf("%s: no Help in %v", name, labels(menu))
		}
		help := menu[index]
		if help.Action != "" || !slices.Equal(labels(help.Children), []string{"About", "Licence"}) ||
			help.Children[0].Action != ActionAbout || help.Children[1].Action != ActionLicence {
			t.Errorf("%s: %+v", name, help)
		}
	}
}

// FR-507.
func TestCloseRequestHidesRatherThanQuits(t *testing.T) {
	t.Parallel()
	r := newRig(t, settings.Defaults())
	if got := r.service.CloseRequested(); got != ActionHide {
		t.Errorf("got %s", got)
	}
}
