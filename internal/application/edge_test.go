package application

import (
	"errors"
	"testing"

	"github.com/oernster/timeribbon/internal/domain/placement"
	"github.com/oernster/timeribbon/internal/domain/settings"
)

// FR-408: two vertical digital cells, 160 + 16 across and 2 x 90 + 16 = 196 along, put against the
// left then the right edge of the primary: flush, centred top to bottom, the place saved.
func TestToEdgePutsAVerticalStripFlushAndKeepsIt(t *testing.T) {
	t.Parallel()
	r := newRig(t, draggedTo(2, settings.Vertical, placement.Point{X: 700, Y: 40}))
	centredY := (1032 - 196) / 2
	for edge, want := range map[placement.Edge]placement.Point{
		placement.Left:  {X: 0, Y: centredY},
		placement.Right: {X: 1920 - 176, Y: centredY},
	} {
		got, err := r.service.ToEdge(placement.Point{X: 700, Y: 40}, edge)
		if err != nil {
			t.Fatal(err)
		}
		if got.At != want || got.Size != (placement.Size{Width: 176, Height: 196}) {
			t.Errorf("%s: got %+v, want at %+v", edge, got, want)
		}
		if saved := r.store.last(t).Placement; saved == nil || saved.Device != primaryMonitor.Device || saved.Offset != want {
			t.Errorf("%s: saved %+v, want offset %+v on the primary", edge, saved, want)
		}
	}
}

// FR-408, FR-407: a horizontal strip on the secondary at 150 percent, 336 x 106 DIP drawn as
// 504 x 159 pixels, goes against that display's top and bottom, centred left to right on it.
func TestToEdgeUsesTheDisplayTheStripIsOn(t *testing.T) {
	t.Parallel()
	at := placement.Point{X: 2500, Y: 100}
	r := newRig(t, clocks(2))
	centredX := 1920 + (2560-504)/2
	for edge, want := range map[placement.Edge]placement.Point{
		placement.Top:    {X: centredX, Y: 0},
		placement.Bottom: {X: centredX, Y: 1392 - 159},
	} {
		got, err := r.service.ToEdge(at, edge)
		if err != nil {
			t.Fatal(err)
		}
		if got.At != want || got.DPI != secondaryMonitor.DPI {
			t.Errorf("%s: got %+v, want at %+v at 144 DPI", edge, got, want)
		}
		if saved := r.store.last(t).Placement; saved == nil || saved.Device != secondaryMonitor.Device {
			t.Errorf("%s: saved %+v, want the secondary", edge, saved)
		}
	}
}

// FR-408, FR-707: a place that cannot be saved raises a notice, one more cell; the strip is fitted
// to hold it and stays flush: 3 x 160 + 16 = 496 along, 106 across, on the bottom edge.
func TestToEdgeThatCannotBeSavedMakesRoomForItsNotice(t *testing.T) {
	t.Parallel()
	r := newRig(t, clocks(2))
	r.store.saveErr = errPlanted
	got, err := r.service.ToEdge(placement.Point{X: 700, Y: 40}, placement.Bottom)
	if err != nil {
		t.Fatal(err)
	}
	want := Arrangement{
		At:   placement.Point{X: (1920 - 496) / 2, Y: 1032 - 106},
		Size: placement.Size{Width: 496, Height: 106}, DPI: placement.BaseDPI,
	}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

// FR-610, FR-408: a strip against the right or bottom edge stays against it when its cells shrink,
// whether it is placed again as a panel closes or refitted where it stands; the left and top edges
// hold its corner, so they keep it anyway. Small vertical digital cells are 120 + 16 = 136 across,
// small horizontal ones 60 + 16 = 76.
func TestShrinkingKeepsTheStripAgainstItsEdge(t *testing.T) {
	t.Parallel()
	cases := []struct {
		orientation settings.Orientation
		edge        placement.Edge
		flush       func(Arrangement) bool
	}{
		{settings.Vertical, placement.Right, func(a Arrangement) bool { return a.At.X+a.Size.Width == 1920 }},
		{settings.Horizontal, placement.Bottom, func(a Arrangement) bool { return a.At.Y+a.Size.Height == 1032 }},
		{settings.Vertical, placement.Left, func(a Arrangement) bool { return a.At.X == 0 }},
		{settings.Horizontal, placement.Top, func(a Arrangement) bool { return a.At.Y == 0 }},
	}
	for _, each := range cases {
		for name, replace := range map[string]func(r rig, at placement.Point) (Arrangement, error){
			"launch":    func(r rig, _ placement.Point) (Arrangement, error) { return r.service.Launch() },
			"rearrange": func(r rig, at placement.Point) (Arrangement, error) { return r.service.Rearrange(at) },
		} {
			r := newRig(t, draggedTo(2, each.orientation, placement.Point{X: 700, Y: 40}))
			if _, err := r.service.Launch(); err != nil {
				t.Fatal(err)
			}
			flush, err := r.service.ToEdge(placement.Point{X: 700, Y: 40}, each.edge)
			if err != nil || !each.flush(flush) {
				t.Fatalf("%s: not against the edge to begin with: %+v %v", each.edge, flush, err)
			}
			if err := r.service.SetSize(settings.Small); err != nil {
				t.Fatal(err)
			}
			got, err := replace(r, flush.At)
			if err != nil {
				t.Fatal(err)
			}
			if !each.flush(got) || got.Size == flush.Size {
				t.Errorf("%s, %s: shrank from %+v to %+v, off its edge", each.edge, name, flush, got)
			}
		}
	}
}

// FR-610: small cells make a smaller strip, two small analogue cells stacked: 120 + 16 across,
// 2 x 100 + 16 along; the snapshot carries the size and its layout, the size is saved; a size the
// setting does not offer is refused.
func TestTheSmallSizeFitsTheStripToSmallCells(t *testing.T) {
	t.Parallel()
	initial := clocks(2)
	initial.Orientation = settings.Vertical
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
	if err := r.service.SetSize("huge"); !errors.Is(err, ErrUnknownChoice) {
		t.Errorf("an unknown size answered %v", err)
	}
}
