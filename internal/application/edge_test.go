package application

import (
	"errors"
	"testing"

	"github.com/oernster/timeribbon/internal/domain/placement"
	"github.com/oernster/timeribbon/internal/domain/settings"
)

// FR-408: two vertical digital cells, 160 + 16 across and 2 x 90 + 16 = 196 along, put against the
// left then the right edge of the primary: flush, centred top to bottom, the place saved.
func TestToEdgePutsAVerticalRibbonFlushAndKeepsIt(t *testing.T) {
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

// FR-408, FR-407: a horizontal ribbon on the secondary at 150 percent, 336 x 106 DIP drawn as
// 504 x 159 pixels, goes against that display's top and bottom, centred left to right on it.
func TestToEdgeUsesTheDisplayTheRibbonIsOn(t *testing.T) {
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

// FR-408, FR-707: a place that cannot be saved raises a notice, one more cell; the ribbon is fitted
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
		Size: placement.Size{Width: 496, Height: 106}, DPI: placement.BaseDPI, Edge: placement.Bottom,
	}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

// FR-410, its acceptance: a vertical ribbon dropped 10 short of the right edge is put flush with its
// top kept, that place stored; dropped 20 short, it stays where it was let go.
func TestADropNearAnEdgeSnapsFlushAndIsStored(t *testing.T) {
	t.Parallel()
	vertical := clocks(2)
	vertical.Orientation = settings.Vertical
	r := newRig(t, vertical)
	got, err := r.service.Moved(placement.Point{X: 1920 - 176 - 10, Y: 300})
	if err != nil {
		t.Fatal(err)
	}
	if got.At != (placement.Point{X: 1920 - 176, Y: 300}) || got.Edge != placement.Right {
		t.Errorf("10 short: got %+v", got)
	}
	if stored := r.store.last(t).Placement; stored.Offset != (placement.Point{X: 1920 - 176, Y: 300}) {
		t.Errorf("stored %+v", stored)
	}
	if got, _ := r.service.Moved(placement.Point{X: 1920 - 176 - 20, Y: 300}); got.At.X != 1920-176-20 || got.Edge != "" {
		t.Errorf("20 short: got %+v", got)
	}
}

// FR-410, FR-619, Oliver's report of 2026-09-29: a vertical ribbon dropped at the top edge is not put
// against it and stands against no edge.
func TestAVerticalRibbonNeverSnapsToTheTop(t *testing.T) {
	t.Parallel()
	vertical := clocks(2)
	vertical.Orientation = settings.Vertical
	r := newRig(t, vertical)
	got, err := r.service.Moved(placement.Point{X: 800, Y: 4})
	if err != nil {
		t.Fatal(err)
	}
	if got.At != (placement.Point{X: 800, Y: 4}) || got.Edge != "" {
		t.Errorf("got %+v", got)
	}
}

// FR-411: a flush place is remembered with its display; a place against no edge keeps the last one.
func TestTheLastEdgeIsRemembered(t *testing.T) {
	t.Parallel()
	vertical := clocks(2)
	vertical.Orientation = settings.Vertical
	r := newRig(t, vertical)
	if _, err := r.service.ToEdge(placement.Point{X: 700, Y: 40}, placement.Left); err != nil {
		t.Fatal(err)
	}
	if _, err := r.service.Moved(placement.Point{X: 800, Y: 300}); err != nil {
		t.Fatal(err)
	}
	want := placement.Against{Device: primaryMonitor.Device, Edge: placement.Left}
	if last := r.service.Settings().LastEdge; last == nil || *last != want {
		t.Errorf("remembered %+v, want %+v", last, want)
	}
}

// FR-613: unpinning away from every edge goes to the edge last used, centred along it; with none
// remembered, to the home edge; with its display gone, to the same edge of the display the ribbon is on.
func TestUnpinningAwayFromAnEdgeGoesToTheLastEdge(t *testing.T) {
	t.Parallel()
	vertical := clocks(2)
	vertical.Orientation = settings.Vertical
	middle := placement.Point{X: 800, Y: 300}
	centredTop := (1032 - 196) / 2

	remembered := vertical
	remembered.LastEdge = &placement.Against{Device: primaryMonitor.Device, Edge: placement.Left}
	if got, err := newRig(t, remembered).service.ToLastEdge(middle); err != nil || got.At != (placement.Point{X: 0, Y: centredTop}) || got.Edge != placement.Left {
		t.Errorf("remembered left: got %+v %v", got, err)
	}
	if got, err := newRig(t, vertical).service.ToLastEdge(middle); err != nil || got.At != (placement.Point{X: 1920 - 176, Y: centredTop}) {
		t.Errorf("none remembered: got %+v %v", got, err)
	}
	gone := vertical
	gone.LastEdge = &placement.Against{Device: `\\.\DISPLAY9`, Edge: placement.Left}
	if got, err := newRig(t, gone).service.ToLastEdge(middle); err != nil || got.At != (placement.Point{X: 0, Y: centredTop}) {
		t.Errorf("display gone: got %+v %v", got, err)
	}
	across := vertical
	across.LastEdge = &placement.Against{Device: primaryMonitor.Device, Edge: placement.Top}
	if got, err := newRig(t, across).service.ToLastEdge(middle); err != nil || got.Edge != placement.Right {
		t.Errorf("remembered edge across the orientation: got %+v %v", got, err)
	}
}

// FR-613: the remembered display, when present, is used even where the ribbon stands on another.
func TestUnpinningGoesToTheRememberedDisplay(t *testing.T) {
	t.Parallel()
	vertical := clocks(2)
	vertical.Orientation = settings.Vertical
	vertical.LastEdge = &placement.Against{Device: secondaryMonitor.Device, Edge: placement.Right}
	got, err := newRig(t, vertical).service.ToLastEdge(placement.Point{X: 800, Y: 300})
	if err != nil {
		t.Fatal(err)
	}
	if got.Edge != placement.Right || got.At.X+got.Size.Width != secondaryMonitor.Work.Right {
		t.Errorf("got %+v, want flush against the secondary's right", got)
	}
}

// FR-610, FR-408: a ribbon against the right or bottom edge stays against it when its cells shrink,
// whether it is placed again as a panel closes or refitted where it stands; the left and top edges
// hold its corner, so they keep it anyway. Small vertical digital cells are 120 + 16 = 136 across,
// small horizontal ones 60 + 16 = 76.
func TestShrinkingKeepsTheRibbonAgainstItsEdge(t *testing.T) {
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

// FR-408, FR-610: a ribbon kept against its edge while it shrinks across its breadth, as turning the
// sun map off shrinks a vertical one, keeps that place across a restart. Measured 2026-09-29: only a
// change of length was saved, so the next launch put the narrower ribbon 16 pixels off the right
// edge, where an unpinned ribbon never collapses (FR-619).
func TestAPlaceKeptAgainstTheEdgeIsSaved(t *testing.T) {
	t.Parallel()
	start := draggedTo(2, settings.Vertical, placement.Point{X: 700, Y: 40})
	start.SunMap = true
	r := newRig(t, start)
	if _, err := r.service.Launch(); err != nil {
		t.Fatal(err)
	}
	flush, err := r.service.ToEdge(placement.Point{X: 700, Y: 40}, placement.Right)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.service.SetSunMap(false); err != nil {
		t.Fatal(err)
	}
	narrower, err := r.service.Rearrange(flush.At)
	if err != nil || narrower.Size.Height != flush.Size.Height || narrower.Size.Width >= flush.Size.Width {
		t.Fatalf("turning the map off gave %+v from %+v (%v); want it narrower and as long", narrower, flush, err)
	}
	restarted := newRig(t, r.store.saved[len(r.store.saved)-1])
	got, err := restarted.service.Launch()
	if err != nil {
		t.Fatal(err)
	}
	if got.Edge != placement.Right || got.At.X+got.Size.Width != 1920 {
		t.Errorf("after a restart the shrunk ribbon stands at %+v, off the right edge", got)
	}
}

// FR-610: small cells make a smaller ribbon, two small analogue cells stacked: 120 + 16 across,
// 2 x 100 + 16 along; the snapshot carries the size and its layout, the size is saved; a size the
// setting does not offer is refused.
func TestTheSmallSizeFitsTheRibbonToSmallCells(t *testing.T) {
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
