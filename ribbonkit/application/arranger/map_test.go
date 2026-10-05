package arranger

import (
	"testing"

	"github.com/oernster/timeribbon/ribbonkit/domain/placement"
	"github.com/oernster/timeribbon/ribbonkit/domain/ribbon"
)

// FR-902, FR-904: a horizontal ribbon at the top edge, its map on and pulled out, has a map below
// it, 480 by 240 for its 336 long ribbon, centred on it; with no map, no side and no map.
func TestAHorizontalRibbonsMapGoesBelowIt(t *testing.T) {
	t.Parallel()
	got, err := newRig(horizontal(), mapPulledOut(2)).arranger.Launch()
	if err != nil {
		t.Fatal(err)
	}
	depth := digitalCell.Height + 2*testPadding + testLane
	want := placement.Rect{Left: got.At.X + (336-480)/2, Top: depth, Right: got.At.X + (336-480)/2 + 480, Bottom: depth + 240}
	if got.MapSide != placement.Bottom || got.Map != want {
		t.Errorf("got side %s map %+v, want bottom %+v", got.MapSide, got.Map, want)
	}
	if off, _ := newRig(horizontal(), cells(2)).arranger.Launch(); off.MapSide != "" || off.Map != (placement.Rect{}) {
		t.Errorf("off: %+v", off)
	}
}

// FR-903, Amendment 22: a horizontal ribbon at the top edge has its handle below it; its map shows
// there only while the pull out is open.
func TestAHorizontalRibbonsMapWaitsForThePullOut(t *testing.T) {
	t.Parallel()
	closed := mapPulledOut(2)
	closed.PullOut = false
	got, err := newRig(horizontal(), closed).arranger.Launch()
	if err != nil {
		t.Fatal(err)
	}
	if got.MapSide != placement.Bottom || got.Map != (placement.Rect{}) {
		t.Errorf("closed: %+v", got)
	}
}

// FR-903, Amendment 23: the content's lane deepens the ribbon whichever way it runs, so a handle
// standing in it covers no cell; its length is unchanged.
func TestTheHandlesLaneDeepensTheRibbon(t *testing.T) {
	t.Parallel()
	for _, orientation := range []ribbon.Orientation{ribbon.Horizontal, ribbon.Vertical} {
		choices := ribbon.Defaults()
		choices.Orientation = orientation
		without, err := newRig(choices, cells(2)).arranger.Launch()
		if err != nil {
			t.Fatal(err)
		}
		with, err := newRig(choices, mapPulledOut(2)).arranger.Launch()
		if err != nil {
			t.Fatal(err)
		}
		wantDeeper := placement.Size{Width: without.Size.Width, Height: without.Size.Height + testLane}
		if orientation == ribbon.Vertical {
			wantDeeper = placement.Size{Width: without.Size.Width + testLane, Height: without.Size.Height}
		}
		if with.Size != wantDeeper {
			t.Errorf("%s: got %+v, want %+v", orientation, with.Size, wantDeeper)
		}
	}
}

// FR-903: a vertical ribbon at the right edge has its handle on the left; its map shows there only
// while the pull out is open.
func TestAVerticalRibbonsMapWaitsForThePullOut(t *testing.T) {
	t.Parallel()
	closed := mapPulledOut(2)
	closed.PullOut = false
	got, err := newRig(vertical(), closed).arranger.Launch()
	if err != nil {
		t.Fatal(err)
	}
	if got.MapSide != placement.Left || got.Map != (placement.Rect{}) {
		t.Errorf("closed: %+v", got)
	}
	got, err = newRig(vertical(), mapPulledOut(2)).arranger.Launch()
	if err != nil {
		t.Fatal(err)
	}
	if got.MapSide != placement.Left || got.Map.Right != got.At.X || got.Map.Width() != 480 {
		t.Errorf("open: %+v", got)
	}
}
