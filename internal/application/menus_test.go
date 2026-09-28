package application

import (
	"slices"
	"testing"

	"github.com/oernster/timeribbon/internal/domain/placement"
	"github.com/oernster/timeribbon/internal/domain/settings"
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
	if !slices.Equal(labels(shown), []string{"Hide strip", "Add clock", "Settings", "Style", "Orientation", "Position", "Always on top", "Help", "Exit"}) ||
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
	item := find(t, r.service.TrayMenu(true), labelAlwaysOnTop)
	if !item.Checkable || item.Checked || item.Action != ActionAlwaysOnTop {
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
func find(t *testing.T, menu []MenuItem, label string) MenuItem {
	t.Helper()
	index := slices.IndexFunc(menu, func(item MenuItem) bool { return item.Label == label })
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
	initial.Orientation = settings.Horizontal
	r := newRig(t, initial)
	for name, menu := range map[string][]MenuItem{"tray": r.service.TrayMenu(true), "context": r.service.ContextMenu()} {
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
	for action, want := range orientationActions {
		if got, ok := OrientationOf(action); !ok || got != want {
			t.Errorf("%s: got %s, %v", action, got, ok)
		}
	}
	if _, ok := StyleOf(ActionVertical); ok {
		t.Error("Vertical was taken for a style")
	}
	if _, ok := OrientationOf(ActionDigital); ok {
		t.Error("Digital was taken for an orientation")
	}
}

// FR-108.
func TestContextMenuOffersTheStripsActions(t *testing.T) {
	t.Parallel()
	r := newRig(t, settings.Defaults())
	if got := labels(r.service.ContextMenu()); !slices.Equal(got, []string{"Add clock", "Settings", "Style", "Orientation", "Position", "Always on top", "Help", "Hide strip", "Exit"}) {
		t.Errorf("got %v", got)
	}
	if last := r.service.ContextMenu()[len(r.service.ContextMenu())-1]; last.Action != ActionExit {
		t.Errorf("the last item acts as %q, want exit", last.Action)
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

// FR-408: Position offers the two edges the strip runs along, in both menus; each item names an
// edge and nothing else does.
func TestPositionOffersTheEdgesAlongTheOrientation(t *testing.T) {
	t.Parallel()
	want := map[settings.Orientation][]MenuAction{
		settings.Vertical:   {ActionLeftEdge, ActionRightEdge},
		settings.Horizontal: {ActionTopEdge, ActionBottomEdge},
	}
	for orientation, actions := range want {
		initial := settings.Defaults()
		initial.Orientation = orientation
		r := newRig(t, initial)
		for name, menu := range map[string][]MenuItem{"tray": r.service.TrayMenu(true), "context": r.service.ContextMenu()} {
			index := slices.IndexFunc(menu, func(item MenuItem) bool { return item.Label == "Position" })
			if index < 0 {
				t.Fatalf("%s %s: no Position in %v", orientation, name, labels(menu))
			}
			var got []MenuAction
			for _, child := range menu[index].Children {
				got = append(got, child.Action)
			}
			if menu[index].Action != "" || !slices.Equal(got, actions) {
				t.Errorf("%s %s: %+v, want %v", orientation, name, menu[index], actions)
			}
		}
	}
	edges := map[MenuAction]placement.Edge{
		ActionLeftEdge: placement.Left, ActionRightEdge: placement.Right,
		ActionTopEdge: placement.Top, ActionBottomEdge: placement.Bottom,
	}
	for action, edge := range edges {
		if got, ok := EdgeOf(action); !ok || got != edge {
			t.Errorf("%s: got %s, %v; want %s", action, got, ok, edge)
		}
	}
	if _, ok := EdgeOf(ActionSettings); ok {
		t.Error("Settings was taken for an edge")
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
