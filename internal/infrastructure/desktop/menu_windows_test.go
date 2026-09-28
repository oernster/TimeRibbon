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

// FR-401: the threshold is Windows' own drag rectangle at 100 percent, which is never zero.
func TestTheDragThresholdIsWindowsOwn(t *testing.T) {
	t.Parallel()
	got := DragThreshold()
	if got.Width <= 0 || got.Height <= 0 {
		t.Errorf("got %+v", got)
	}
	t.Logf("drag threshold %+v DIP", got)
}
