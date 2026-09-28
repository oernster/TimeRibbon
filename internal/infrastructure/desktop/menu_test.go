package desktop

import (
	"testing"

	"github.com/oernster/timeribbon/internal/application"
)

func TestTheChosenIdentifierNamesItsItem(t *testing.T) {
	t.Parallel()
	items := []application.MenuItem{{Action: application.ActionHide}, {Action: application.ActionExit}}
	cases := map[int]struct {
		action application.MenuAction
		ok     bool
	}{
		0:              {"", false},
		menuIDBase:     {application.ActionHide, true},
		menuIDBase + 1: {application.ActionExit, true},
		menuIDBase + 2: {"", false},
	}
	for id, want := range cases {
		if action, ok := chosenAction(items, id); action != want.action || ok != want.ok {
			t.Errorf("id %d: got %q %v", id, action, ok)
		}
	}
}

// FR-508: a submenu takes no identifier of its own; its items are numbered after every item before
// it and before every item after it.
func TestASubmenuIsNumberedAfterEveryItemBeforeIt(t *testing.T) {
	t.Parallel()
	items := []application.MenuItem{
		{Action: application.ActionSettings},
		{Label: "Help", Children: []application.MenuItem{{Action: application.ActionAbout}, {Action: application.ActionLicence}}},
		{Action: application.ActionExit},
	}
	want := []application.MenuAction{application.ActionSettings, application.ActionAbout, application.ActionLicence, application.ActionExit}
	for offset, action := range want {
		if got, ok := chosenAction(items, menuIDBase+offset); !ok || got != action {
			t.Errorf("id %d: got %q %v, want %q", menuIDBase+offset, got, ok, action)
		}
	}
	if _, ok := chosenAction(items, menuIDBase+len(want)); ok {
		t.Error("an identifier past the last item named one")
	}
}

// FR-401: the threshold is Windows' own drag rectangle at 100 percent, which is never zero.
func TestTheDragThresholdIsWindowsOwn(t *testing.T) {
	t.Parallel()
	got := DragThreshold()
	if got.Width <= 0 || got.Height <= 0 {
		t.Errorf("got %+v", got)
	}
	t.Logf("drag threshold %+v DIP", got)
}
