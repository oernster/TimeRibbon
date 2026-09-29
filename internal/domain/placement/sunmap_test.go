package placement

import "testing"

// work is a display 1920 by 1032 at 100 percent, so DIP and pixels agree.
var work = Rect{Left: 0, Top: 0, Right: 1920, Bottom: 1032}

// FR-902, FR-903: the map goes on the side away from the ribbon's edge; against no edge, the side
// with more room; at an equal room, below a horizontal ribbon and left of a vertical one.
func TestTheMapAdjoinsTheSideAwayFromTheEdge(t *testing.T) {
	t.Parallel()
	horizontal := Rect{Left: 700, Top: 0, Right: 1036, Bottom: 106}
	vertical := Rect{Left: 1744, Top: 400, Right: 1920, Bottom: 596}
	cases := []struct {
		name     string
		ribbon   Rect
		vertical bool
		edge     Edge
		want     Edge
	}{
		{"at the top", horizontal, false, Top, Bottom},
		{"at the bottom", horizontal, false, Bottom, Top},
		{"at the right", vertical, true, Right, Left},
		{"at the left", vertical, true, Left, Right},
		{"horizontal, more room above", Rect{Left: 700, Top: 800, Right: 1036, Bottom: 906}, false, "", Top},
		{"horizontal, more room below", Rect{Left: 700, Top: 100, Right: 1036, Bottom: 206}, false, "", Bottom},
		{"horizontal, even", Rect{Left: 700, Top: 463, Right: 1036, Bottom: 569}, false, "", Bottom},
		{"vertical, more room right", Rect{Left: 100, Top: 400, Right: 276, Bottom: 596}, true, "", Right},
		{"vertical, more room left", Rect{Left: 1500, Top: 400, Right: 1676, Bottom: 596}, true, "", Left},
		{"vertical, even", Rect{Left: 872, Top: 400, Right: 1048, Bottom: 596}, true, "", Left},
	}
	for _, each := range cases {
		if got := InnerSide(each.ribbon, work, each.vertical, each.edge); got != each.want {
			t.Errorf("%s: got %s, want %s", each.name, got, each.want)
		}
	}
}

// FR-904, its acceptance: a horizontal ribbon 1200 long at the top gets a 1200 by 600 map below it;
// one 336 long gets the 480 by 240 minimum centred on it.
func TestTheMapMatchesTheRibbon(t *testing.T) {
	t.Parallel()
	long := Rect{Left: 360, Top: 0, Right: 1560, Bottom: 106}
	if got, ok := MapBeside(long, work, Bottom, MapMinimumWidth, MapFloor); !ok || got != (Rect{Left: 360, Top: 106, Right: 1560, Bottom: 706}) {
		t.Errorf("1200 long: got %+v %v", got, ok)
	}
	short := Rect{Left: 792, Top: 0, Right: 1128, Bottom: 106}
	if got, ok := MapBeside(short, work, Bottom, MapMinimumWidth, MapFloor); !ok || got != (Rect{Left: 720, Top: 106, Right: 1200, Bottom: 346}) {
		t.Errorf("336 long: got %+v %v", got, ok)
	}
	above := Rect{Left: 792, Top: 926, Right: 1128, Bottom: 1032}
	if got, ok := MapBeside(above, work, Top, MapMinimumWidth, MapFloor); !ok || got != (Rect{Left: 720, Top: 686, Right: 1200, Bottom: 926}) {
		t.Errorf("above: got %+v %v", got, ok)
	}
}

// FR-904, its acceptance: a vertical ribbon 1032 long with 700 of room beside it gets a 700 by 350
// map centred on it; a horizontal ribbon with 300 of room below gets a 600 by 300 map.
func TestTheMapScalesToTheRoom(t *testing.T) {
	t.Parallel()
	vertical := Rect{Left: 700, Top: 0, Right: 876, Bottom: 1032}
	if got, ok := MapBeside(vertical, work, Left, MapMinimumWidth, MapFloor); !ok || got != (Rect{Left: 0, Top: 341, Right: 700, Bottom: 691}) {
		t.Errorf("vertical: got %+v %v", got, ok)
	}
	leftEdge := Rect{Left: 0, Top: 0, Right: 176, Bottom: 1032}
	if got, ok := MapBeside(leftEdge, work, Right, MapMinimumWidth, MapFloor); !ok || got != (Rect{Left: 176, Top: 258, Right: 1208, Bottom: 774}) {
		t.Errorf("vertical at the left: got %+v %v", got, ok)
	}
	low := Rect{Left: 360, Top: 626, Right: 1560, Bottom: 732}
	if got, ok := MapBeside(low, work, Bottom, MapMinimumWidth, MapFloor); !ok || got != (Rect{Left: 660, Top: 732, Right: 1260, Bottom: 1032}) {
		t.Errorf("horizontal: got %+v %v", got, ok)
	}
}

// FR-904: under 120 of room no map is shown; a map near the work area's end is kept inside it.
func TestTooLittleRoomShowsNoMap(t *testing.T) {
	t.Parallel()
	cramped := Rect{Left: 0, Top: 920, Right: 336, Bottom: 1026}
	if got, ok := MapBeside(cramped, work, Bottom, MapMinimumWidth, MapFloor); ok {
		t.Errorf("6 of room: got %+v", got)
	}
	corner := Rect{Left: 0, Top: 0, Right: 336, Bottom: 106}
	if got, ok := MapBeside(corner, work, Bottom, MapMinimumWidth, MapFloor); !ok || got.Left != 0 || got.Width() != MapMinimumWidth {
		t.Errorf("in the corner: got %+v %v", got, ok)
	}
	right := Rect{Left: 1744, Top: 0, Right: 1920, Bottom: 196}
	if got, ok := MapBeside(right, work, Left, MapMinimumWidth, MapFloor); !ok || got.Top != 0 || got.Right != 1744 {
		t.Errorf("vertical at the top: got %+v %v", got, ok)
	}
}
