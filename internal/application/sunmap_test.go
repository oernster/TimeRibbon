package application

import (
	"testing"

	"github.com/oernster/timeribbon/internal/domain/settings"
	"github.com/oernster/timeribbon/internal/domain/sun"
	"github.com/oernster/timeribbon/ribbonkit/domain/placement"
	"github.com/oernster/timeribbon/ribbonkit/domain/ribbon"
)

// londonAt is Europe/London's city as the rig's catalogue gives it.
var londonAt = sun.Point{Latitude: 51.5083, Longitude: -0.1253}

// FR-901: both menus hold Sun map after Pin ribbon, unticked until chosen; the choice is saved.
func TestBothMenusOfferSunMap(t *testing.T) {
	t.Parallel()
	r := newRig(t, settings.Defaults())
	for name, menu := range map[string][]MenuItem{"tray": r.service.TrayMenu(true), "context": r.service.ContextMenu()} {
		if item := find(t, menu, labelSunMap); item.Action != ActionSunMap || !item.Checkable || item.Checked {
			t.Errorf("%s: %+v", name, item)
		}
	}
	if err := r.service.SetSunMap(true); err != nil {
		t.Fatal(err)
	}
	if !find(t, r.service.ContextMenu(), labelSunMap).Checked || !r.store.last(t).SunMap {
		t.Error("choosing Sun map did not tick it or save it")
	}
	if err := r.service.SetPullOut(true); err != nil || !r.store.last(t).PullOut {
		t.Errorf("the pull out was not saved: %v", err)
	}
}

// FR-908: a clock whose zone has a city is marked there with its label; one whose zone has none
// (UTC) and one that cannot be shown are not.
func TestAZoneWithNoPlaceHasNoMark(t *testing.T) {
	t.Parallel()
	initial := settings.Defaults()
	initial.SunMap = true
	initial = initial.WithClockAdded(settings.Entry{ID: "a", Zone: "Europe/London", Label: "Mum"})
	initial = initial.WithClockAdded(settings.Entry{ID: "b", Zone: "UTC", Label: "Server"})
	initial = initial.WithClockAdded(settings.Entry{ID: "c", Zone: "Not/AZone", Label: "Nowhere"})
	got := newRig(t, initial).service.Snapshot().SunMap
	if !got.On || len(got.Marks) != 1 || got.Marks[0] != (Mark{Label: "Mum", At: londonAt}) {
		t.Errorf("got %+v", got)
	}
}

// FR-905: the snapshot carries the subsolar point at its own instant; with no clocks, no marks.
func TestTheSnapshotCarriesTheSubsolarPoint(t *testing.T) {
	t.Parallel()
	r := newRig(t, settings.Defaults())
	snapshot := r.service.Snapshot()
	if snapshot.SunMap.Subsolar != sun.Subsolar(snapshot.Now) || snapshot.SunMap.Marks == nil || len(snapshot.SunMap.Marks) != 0 {
		t.Errorf("got %+v", snapshot.SunMap)
	}
}

// FR-902, FR-904: a horizontal ribbon at the top edge, sun map on and pulled out, has a map below it,
// 480 by 240 for its 336 long ribbon, centred on it; off, no side and no map.
func TestAHorizontalRibbonsMapGoesBelowIt(t *testing.T) {
	t.Parallel()
	on := clocks(2)
	on.SunMap, on.PullOut = true, true
	got, err := newRig(t, on).service.Launch()
	if err != nil {
		t.Fatal(err)
	}
	depth := testLayout.Digital.Height + 2*testLayout.Padding + testLayout.HandleLane
	want := placement.Rect{Left: got.At.X + (336-480)/2, Top: depth, Right: got.At.X + (336-480)/2 + 480, Bottom: depth + 240}
	if got.MapSide != placement.Bottom || got.Map != want {
		t.Errorf("got side %s map %+v, want bottom %+v", got.MapSide, got.Map, want)
	}
	if off, _ := newRig(t, clocks(2)).service.Launch(); off.MapSide != "" || off.Map != (placement.Rect{}) {
		t.Errorf("off: %+v", off)
	}
}

// FR-903, Amendment 22: a horizontal ribbon at the top edge has its handle below it; its map shows
// there only while the pull out is open.
func TestAHorizontalRibbonsMapWaitsForThePullOut(t *testing.T) {
	t.Parallel()
	closed := clocks(2)
	closed.SunMap = true
	got, err := newRig(t, closed).service.Launch()
	if err != nil {
		t.Fatal(err)
	}
	if got.MapSide != placement.Bottom || got.Map != (placement.Rect{}) {
		t.Errorf("closed: %+v", got)
	}
}

// FR-903, Amendment 23: while the sun map is on, the ribbon is deeper by the handle's lane, open or
// closed and whichever way it runs, so the handle covers no cell; its length is unchanged.
func TestTheHandlesLaneDeepensTheRibbon(t *testing.T) {
	t.Parallel()
	for _, orientation := range []ribbon.Orientation{ribbon.Horizontal, ribbon.Vertical} {
		for _, pullOut := range []bool{false, true} {
			off := clocks(2)
			off.Orientation = orientation
			on := off
			on.SunMap, on.PullOut = true, pullOut
			without, err := newRig(t, off).service.Launch()
			if err != nil {
				t.Fatal(err)
			}
			with, err := newRig(t, on).service.Launch()
			if err != nil {
				t.Fatal(err)
			}
			wantDeeper := placement.Size{Width: without.Size.Width, Height: without.Size.Height + testLayout.HandleLane}
			if orientation == ribbon.Vertical {
				wantDeeper = placement.Size{Width: without.Size.Width + testLayout.HandleLane, Height: without.Size.Height}
			}
			if with.Size != wantDeeper {
				t.Errorf("%s, pull out %v: got %+v, want %+v", orientation, pullOut, with.Size, wantDeeper)
			}
		}
	}
}

// FR-903: a vertical ribbon at the right edge has its handle on the left; its map shows there only
// while the pull out is open.
func TestAVerticalRibbonsMapWaitsForThePullOut(t *testing.T) {
	t.Parallel()
	closed := clocks(2)
	closed.Orientation, closed.SunMap = ribbon.Vertical, true
	got, err := newRig(t, closed).service.Launch()
	if err != nil {
		t.Fatal(err)
	}
	if got.MapSide != placement.Left || got.Map != (placement.Rect{}) {
		t.Errorf("closed: %+v", got)
	}
	open := closed
	open.PullOut = true
	got, err = newRig(t, open).service.Launch()
	if err != nil {
		t.Fatal(err)
	}
	if got.MapSide != placement.Left || got.Map.Right != got.At.X || got.Map.Width() != 480 {
		t.Errorf("open: %+v", got)
	}
}
