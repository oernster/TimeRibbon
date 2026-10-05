package application

import (
	"slices"
	"testing"

	"github.com/oernster/ribbonkit/application/menus"
	"github.com/oernster/ribbonkit/domain/ribbon"
	"github.com/oernster/timeribbon/internal/domain/settings"
)

// commandLabels are the menu items that choose nothing, so Settings does not offer them (FR-624).
var commandLabels = []string{labelShow, labelHide, labelAddClock, labelSettings, labelHelp, labelExit}

// FR-624: every item of either menu is a command or one of Settings' choices, so a choice added to
// the menus and not to Settings fails here.
func TestEveryMenuChoiceIsOfferedBySettings(t *testing.T) {
	t.Parallel()
	r := newRig(t, settings.Defaults())
	offered := labels(r.service.SettingsChoices())
	for _, menu := range [][]menus.Item{r.service.TrayMenu(true), r.service.TrayMenu(false), r.service.ContextMenu()} {
		for _, item := range menu {
			if !slices.Contains(offered, item.Label) && !slices.Contains(commandLabels, item.Label) {
				t.Errorf("%s is on a menu and not in Settings", item.Label)
			}
		}
	}
	for _, label := range commandLabels {
		if slices.Contains(offered, label) {
			t.Errorf("Settings offers the command %s", label)
		}
	}
}

// FR-624: Settings' choices are the menus' own items, ticks included, so the two cannot disagree.
func TestSettingsChoicesAreTheMenusOwnItems(t *testing.T) {
	t.Parallel()
	r := newRig(t, settings.Defaults())
	if err := r.service.SetColour(ribbon.Ocean); err != nil {
		t.Fatal(err)
	}
	menu := r.service.ContextMenu()
	for _, choice := range r.service.SettingsChoices() {
		if item := find(t, menu, choice.Label); !slices.Equal(labels(item.Children), labels(choice.Children)) ||
			item.Checked != choice.Checked || item.Action != choice.Action {
			t.Errorf("Settings' %s is %+v, the menu's %+v", choice.Label, choice, item)
		}
	}
	colour := find(t, r.service.SettingsChoices(), labelColour)
	if ocean := find(t, colour.Children, colourLabels[ribbon.Ocean]); !ocean.Checked {
		t.Errorf("Ocean chosen and not ticked: %+v", ocean)
	}
}
