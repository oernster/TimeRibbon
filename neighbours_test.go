package main

import (
	"slices"
	"strings"
	"testing"

	"github.com/oernster/timeribbon/ribbonkit/domain/identity"
	"github.com/oernster/timeribbon/ribbonkit/domain/placement"
	"github.com/oernster/timeribbon/ribbonkit/infrastructure/occupancy"
)

// FR-412: TimeRibbon takes its place in the folder the environment names, where another product
// opened there sees what it holds; an environment naming none is said in the log and is alone.
func TestTimeRibbonTakesItsPlaceAmongTheRibbons(t *testing.T) {
	t.Parallel()
	folder := t.TempDir()
	lookup := func(string) (string, bool) { return folder, true }
	var log strings.Builder
	ours := openNeighbours(lookup, &log)
	defer ours.Close()
	shared, err := occupancy.Dir(lookup)
	if err != nil {
		t.Fatal(err)
	}
	other := occupancy.Open(shared, identity.App{Name: "OtherRibbon", AppID: "uk.example.OtherRibbon"}, &log)
	defer other.Close()
	held := []placement.Rect{{Left: 1744, Top: 418, Right: 1920, Bottom: 614}}
	ours.Hold(held)
	if got := other.Taken(); !slices.Equal(got, held) || log.Len() != 0 {
		t.Errorf("the other product saw %+v, logged %q; want TimeRibbon's %+v", got, log.String(), held)
	}

	var unnamed strings.Builder
	alone := openNeighbours(func(string) (string, bool) { return "", false }, &unnamed)
	defer alone.Close()
	if got := alone.Taken(); len(got) != 0 || !strings.Contains(unnamed.String(), "keeping off no other ribbon") {
		t.Errorf("with no folder named: saw %+v, logged %q", got, unnamed.String())
	}
}
