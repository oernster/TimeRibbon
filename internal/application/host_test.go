package application

import (
	"errors"
	"testing"

	"github.com/oernster/timeribbon/internal/domain/settings"
	"github.com/oernster/timeribbon/ribbonkit/application/arranger"
	"github.com/oernster/timeribbon/ribbonkit/domain/placement"
	"github.com/oernster/timeribbon/ribbonkit/domain/ribbon"
)

// clocks answers horizontal settings holding n London clocks; the arithmetic below is worked for
// horizontal cells, whatever the default orientation is.
func clocks(n int) settings.Settings {
	s := settings.Defaults()
	s.Orientation = ribbon.Horizontal
	for range n {
		s = s.WithClockAdded(settings.Entry{ID: "x", Zone: "Europe/London"})
	}
	return s
}

// FR-105, FR-107, FR-603, FR-604, FR-903: the arranger is handed the style's cell, the prompt's
// when there are no clocks, one cell per clock and the handle's lane while the sun map is on,
// pulled out or not; the map's choices are handed with it.
func TestTheServiceHandsTheArrangerItsContent(t *testing.T) {
	t.Parallel()
	analogue := clocks(3)
	analogue.Style = settings.Analogue
	closed := clocks(2)
	closed.SunMap = true
	open := closed
	open.PullOut = true
	for name, each := range map[string]struct {
		settings settings.Settings
		want     arranger.Content
	}{
		"digital":  {clocks(2), arranger.Content{Cell: testLayout.Digital, Cells: 2, Padding: testLayout.Padding}},
		"analogue": {analogue, arranger.Content{Cell: testLayout.Analogue, Cells: 3, Padding: testLayout.Padding}},
		"empty":    {clocks(0), arranger.Content{Cell: testLayout.Prompt, Cells: 1, Padding: testLayout.Padding}},
		"map":      {closed, arranger.Content{Cell: testLayout.Digital, Cells: 2, Padding: testLayout.Padding, Lane: testLayout.HandleLane, Beside: true}},
		"pull out": {open, arranger.Content{Cell: testLayout.Digital, Cells: 2, Padding: testLayout.Padding, Lane: testLayout.HandleLane, Beside: true, PullOut: true}},
	} {
		r := newRig(t, each.settings)
		choices, got := host{r.service}.Ribbon()
		if got != each.want || choices != r.service.Settings().Choices {
			t.Errorf("%s: handed %+v with %+v, want %+v with the settings' own", name, got, choices, each.want)
		}
	}
}

// FR-707: a change the arranger makes is saved with the rest of the settings, so a save that fails
// raises the same notice.
func TestTheArrangersChangesAreSavedWithTheSettings(t *testing.T) {
	t.Parallel()
	r := newRig(t, clocks(2))
	moved := func(c ribbon.Choices) ribbon.Choices {
		c.LastEdge = &placement.Against{Device: primaryMonitor.Device, Edge: placement.Top}
		return c
	}
	if err := (host{r.service}).ChangeRibbon(moved); err != nil || r.store.last(t).LastEdge == nil || len(r.store.last(t).Clocks) != 2 {
		t.Fatalf("saved %+v (%v), want the edge saved with the clocks", r.store.last(t), err)
	}
	r.store.saveErr = errPlanted
	if err := (host{r.service}).ChangeRibbon(moved); !errors.Is(err, errPlanted) || len(r.service.Snapshot().Notices) != 1 {
		t.Errorf("a failed save answered %v with notices %v", err, r.service.Snapshot().Notices)
	}
}

// FR-107: an empty ribbon is sized for its prompt, whatever the style: 160 + 16 by 190 + 16.
func TestAnEmptyRibbonIsSizedForItsPrompt(t *testing.T) {
	t.Parallel()
	r := newRig(t, clocks(0))
	got, _ := r.service.Launch()
	if got.Size != (placement.Size{Width: 176, Height: 206}) {
		t.Errorf("got %+v", got.Size)
	}
}

// FR-106, FR-707: a notice is drawn as one more cell, so the ribbon makes room for it; two clocks and
// a notice are 3 x 160 + 2 x 8 = 496 along. Once the notice is dismissed the ribbon is 336 again.
func TestTheRibbonMakesRoomForANotice(t *testing.T) {
	t.Parallel()
	r := newRig(t, clocks(2))
	r.store.saveErr = errPlanted
	if err := r.service.SetTheme(ribbon.Dark); !errors.Is(err, errPlanted) {
		t.Fatalf("the save did not fail: %v", err)
	}
	got, _ := r.service.Launch()
	if got.Size.Width != 496 || got.Scrolls {
		t.Errorf("with a notice: got %+v", got)
	}
	r.store.saveErr = nil
	r.service.DismissNotices()
	got, _ = r.service.Launch()
	if got.Size.Width != 336 {
		t.Errorf("after dismissing: got %+v", got)
	}
}

// FR-610: small cells make a smaller ribbon, two small analogue cells stacked: 120 + 16 across,
// 2 x 100 + 16 along; the snapshot carries the size and its layout, the size is saved; a size the
// setting does not offer is refused.
func TestTheSmallSizeFitsTheRibbonToSmallCells(t *testing.T) {
	t.Parallel()
	initial := clocks(2)
	initial.Orientation = ribbon.Vertical
	initial.Style = settings.Analogue
	r := newRig(t, initial)
	if err := r.service.SetSize(settings.Small); err != nil {
		t.Fatal(err)
	}
	if got, _ := r.service.Launch(); got.Size != (placement.Size{Width: 136, Height: 216}) {
		t.Errorf("got %+v", got.Size)
	}
	if snapshot := r.service.Snapshot(); snapshot.Size != settings.Small || snapshot.Layout != testLayouts.Small {
		t.Errorf("the snapshot carries %q and %+v, want small", snapshot.Size, snapshot.Layout)
	}
	if saved := r.store.last(t); saved.Size != settings.Small {
		t.Errorf("saved size %q, want small", saved.Size)
	}
	if err := r.service.SetSize("huge"); !errors.Is(err, ribbon.ErrUnknownChoice) {
		t.Errorf("an unknown size answered %v", err)
	}
}
