package desktop

import (
	"testing"

	"github.com/oernster/timeribbon/internal/application"
)

// FR-508: a submenu takes no number of its own; its items are numbered after every item before it
// and before every item after it.
func TestASubmenuIsNumberedAfterEveryItemBeforeIt(t *testing.T) {
	t.Parallel()
	items := []application.MenuItem{
		{Action: application.ActionSettings},
		{Label: "Help", Children: []application.MenuItem{{Action: application.ActionAbout}, {Action: application.ActionLicence}}},
		{Action: application.ActionExit},
	}
	want := []application.MenuAction{application.ActionSettings, application.ActionAbout, application.ActionLicence, application.ActionExit}
	for index, action := range want {
		if got, ok := actionAt(items, index); !ok || got != action {
			t.Errorf("index %d: got %q %v, want %q", index, got, ok, action)
		}
	}
	for _, index := range []int{-1, len(want)} {
		if _, ok := actionAt(items, index); ok {
			t.Errorf("index %d named an item", index)
		}
	}
}
