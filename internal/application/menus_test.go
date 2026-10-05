package application

import (
	"errors"
	"slices"
	"testing"

	"github.com/oernster/ribbonkit/application/menus"
	"github.com/oernster/ribbonkit/domain/ribbon"
	"github.com/oernster/timeribbon/internal/domain/settings"
)

func labels(items []menus.Item) []string {
	var out []string
	for _, item := range items {
		out = append(out, item.Label)
	}
	return out
}

// FR-502, FR-613, FR-901.
func TestTrayMenuNamesTheOppositeOfTheVisibility(t *testing.T) {
	t.Parallel()
	r := newRig(t, settings.Defaults())
	shown := r.service.TrayMenu(true)
	if !slices.Equal(labels(shown), []string{"Hide ribbon", "Add clock", "Settings", "Style", "Colour", "Orientation", "Position", "Always on top", "Pin ribbon", "Sun map", "Help", "Exit"}) ||
		shown[0].Action != menus.Hide {
		t.Errorf("visible: %+v", shown)
	}
	hidden := r.service.TrayMenu(false)
	if hidden[0].Label != "Show ribbon" || hidden[0].Action != menus.Show {
		t.Errorf("hidden: %+v", hidden[0])
	}
}

// FR-505, FR-502: the Always on top item shows its state.
func TestAlwaysOnTopItemShowsItsState(t *testing.T) {
	t.Parallel()
	r := newRig(t, settings.Defaults())
	item := find(t, r.service.TrayMenu(true), labelAlwaysOnTop)
	if !item.Checkable || item.Checked || item.Action != menus.AlwaysOnTop {
		t.Errorf("off: %+v", item)
	}
	if err := r.service.SetAlwaysOnTop(true); err != nil {
		t.Fatal(err)
	}
	if item := find(t, r.service.ContextMenu(), labelAlwaysOnTop); !item.Checked {
		t.Errorf("on: %+v", item)
	}
}

// find answers the item of menu labelled label, failing the test when there is none.
func find(t *testing.T, menu []menus.Item, label string) menus.Item {
	t.Helper()
	index := slices.IndexFunc(menu, func(item menus.Item) bool { return item.Label == label })
	if index < 0 {
		t.Fatalf("no %s in %v", label, labels(menu))
	}
	return menu[index]
}

// FR-108, FR-502: Style and Orientation are submenus in both menus, the current choice ticked;
// each item names what it chooses and nothing else does.
func TestBothMenusOfferStyleAndOrientationWithTheCurrentTicked(t *testing.T) {
	t.Parallel()
	initial := settings.Defaults()
	initial.Style = settings.Analogue
	initial.Orientation = ribbon.Horizontal
	r := newRig(t, initial)
	for name, menu := range map[string][]menus.Item{"tray": r.service.TrayMenu(true), "context": r.service.ContextMenu()} {
		style, orientation := find(t, menu, labelStyle), find(t, menu, labelOrientation)
		if style.Action != "" || !slices.Equal(labels(style.Children), []string{"Digital", "Analogue"}) ||
			style.Children[0].Checked || !style.Children[1].Checked || !style.Children[0].Checkable {
			t.Errorf("%s style: %+v", name, style)
		}
		if orientation.Action != "" || !slices.Equal(labels(orientation.Children), []string{"Horizontal", "Vertical"}) ||
			!orientation.Children[0].Checked || orientation.Children[1].Checked || !orientation.Children[1].Checkable {
			t.Errorf("%s orientation: %+v", name, orientation)
		}
	}
	for action, want := range styleActions {
		if got, ok := StyleOf(action); !ok || got != want {
			t.Errorf("%s: got %s, %v", action, got, ok)
		}
	}
	if _, ok := StyleOf(menus.OrientVertical); ok {
		t.Error("Vertical was taken for a style")
	}
	if _, ok := menus.OrientationOf(ActionDigital); ok {
		t.Error("Digital was taken for an orientation")
	}
}

// FR-108, FR-613, FR-901.
func TestContextMenuOffersTheRibbonsActions(t *testing.T) {
	t.Parallel()
	r := newRig(t, settings.Defaults())
	if got := labels(r.service.ContextMenu()); !slices.Equal(got, []string{"Add clock", "Settings", "Style", "Colour", "Orientation", "Position", "Always on top", "Pin ribbon", "Sun map", "Help", "Hide ribbon", "Exit"}) {
		t.Errorf("got %v", got)
	}
	if last := r.service.ContextMenu()[len(r.service.ContextMenu())-1]; last.Action != menus.Exit {
		t.Errorf("the last item acts as %q, want exit", last.Action)
	}
}

// FR-508, FR-509: Help is a submenu with no action of its own, holding About, Licence and Check for
// updates, in both menus.
func TestBothMenusOfferHelpWithAboutLicenceAndUpdates(t *testing.T) {
	t.Parallel()
	r := newRig(t, settings.Defaults())
	for name, menu := range map[string][]menus.Item{"tray": r.service.TrayMenu(true), "context": r.service.ContextMenu()} {
		index := slices.IndexFunc(menu, func(item menus.Item) bool { return item.Label == "Help" })
		if index < 0 {
			t.Fatalf("%s: no Help in %v", name, labels(menu))
		}
		help := menu[index]
		if help.Action != "" || !slices.Equal(labels(help.Children), []string{"About", "Licence", "Check for updates"}) ||
			help.Children[0].Action != menus.About || help.Children[1].Action != menus.Licence ||
			help.Children[2].Action != menus.Updates {
			t.Errorf("%s: %+v", name, help)
		}
	}
}

// FR-408: Position offers the two edges the ribbon runs along, in both menus. Which edge each item
// names is ribbonkit's menus package's test.
func TestPositionOffersTheEdgesAlongTheOrientation(t *testing.T) {
	t.Parallel()
	want := map[ribbon.Orientation][]menus.Action{
		ribbon.Vertical:   {menus.LeftEdge, menus.RightEdge},
		ribbon.Horizontal: {menus.TopEdge, menus.BottomEdge},
	}
	for orientation, actions := range want {
		initial := settings.Defaults()
		initial.Orientation = orientation
		r := newRig(t, initial)
		for name, menu := range map[string][]menus.Item{"tray": r.service.TrayMenu(true), "context": r.service.ContextMenu()} {
			index := slices.IndexFunc(menu, func(item menus.Item) bool { return item.Label == "Position" })
			if index < 0 {
				t.Fatalf("%s %s: no Position in %v", orientation, name, labels(menu))
			}
			var got []menus.Action
			for _, child := range menu[index].Children {
				got = append(got, child.Action)
			}
			if menu[index].Action != "" || !slices.Equal(got, actions) {
				t.Errorf("%s %s: %+v, want %v", orientation, name, menu[index], actions)
			}
		}
	}
}

// FR-611: Colour is a submenu in both menus offering every scheme with the current one ticked;
// each item names its scheme and nothing else does.
func TestBothMenusOfferEveryColourWithTheCurrentTicked(t *testing.T) {
	t.Parallel()
	initial := settings.Defaults()
	initial.Colour = ribbon.Ocean
	r := newRig(t, initial)
	for name, menu := range map[string][]menus.Item{"tray": r.service.TrayMenu(true), "context": r.service.ContextMenu()} {
		colour := find(t, menu, labelColour)
		if colour.Action != "" || !slices.Equal(labels(colour.Children), []string{
			"Classic", "Neon", "Ocean", "Sunset", "Forest", "Amber", "Ruby", "Indigo", "Berry", "Contrast",
		}) {
			t.Fatalf("%s: %+v", name, colour)
		}
		for index, item := range colour.Children {
			got, ok := menus.ColourOf(item.Action)
			if !ok || got != ribbon.Colours[index] || !item.Checkable || item.Checked != (got == ribbon.Ocean) {
				t.Errorf("%s: %+v answered %s, %v", name, item, got, ok)
			}
		}
	}
	for _, other := range []menus.Action{ActionDigital, "colour-mauve", "colour-"} {
		if _, ok := menus.ColourOf(other); ok {
			t.Errorf("%q was taken for a colour", other)
		}
	}
	if err := r.service.SetColour("mauve"); !errors.Is(err, ribbon.ErrUnknownChoice) {
		t.Errorf("an unknown colour answered %v", err)
	}
	if err := r.service.SetColour(ribbon.Neon); err != nil || r.store.last(t).Colour != ribbon.Neon || r.service.Snapshot().Colour != ribbon.Neon {
		t.Errorf("neon was not chosen, saved and shown: %v", err)
	}
}

// FR-507.
func TestCloseRequestHidesRatherThanQuits(t *testing.T) {
	t.Parallel()
	r := newRig(t, settings.Defaults())
	if got := r.service.CloseRequested(); got != menus.Hide {
		t.Errorf("got %s", got)
	}
}

// FR-613: both menus hold Pin ribbon directly after Always on top, ticked while pinned; the choice
// is kept.
func TestBothMenusOfferPinAfterAlwaysOnTop(t *testing.T) {
	t.Parallel()
	r := newRig(t, settings.Defaults())
	for name, menu := range map[string][]menus.Item{"tray": r.service.TrayMenu(true), "context": r.service.ContextMenu()} {
		index := -1
		for position, item := range menu {
			if item.Action == menus.AlwaysOnTop {
				index = position
			}
		}
		if index < 0 || index+1 >= len(menu) {
			t.Fatalf("%s: no item follows Always on top", name)
		}
		if pin := menu[index+1]; pin.Action != menus.Pin || pin.Label != labelPin || !pin.Checkable || !pin.Checked {
			t.Errorf("%s: after Always on top came %+v, want Pin ribbon ticked", name, pin)
		}
	}
	if err := r.service.SetPinned(false); err != nil {
		t.Fatal(err)
	}
	if find(t, r.service.ContextMenu(), labelPin).Checked {
		t.Error("Pin ribbon is still ticked once unpinned")
	}
	if saved := r.store.saved[len(r.store.saved)-1]; saved.Pinned {
		t.Error("unpinning was not saved")
	}
}
