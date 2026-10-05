package application

import (
	"testing"

	"github.com/oernster/timeribbon/internal/domain/settings"
	"github.com/oernster/timeribbon/internal/domain/sun"
	"github.com/oernster/timeribbon/ribbonkit/application/menus"
)

// londonAt is Europe/London's city as the rig's catalogue gives it.
var londonAt = sun.Point{Latitude: 51.5083, Longitude: -0.1253}

// FR-901: both menus hold Sun map after Pin ribbon, unticked until chosen; the choice is saved.
func TestBothMenusOfferSunMap(t *testing.T) {
	t.Parallel()
	r := newRig(t, settings.Defaults())
	for name, menu := range map[string][]menus.Item{"tray": r.service.TrayMenu(true), "context": r.service.ContextMenu()} {
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
