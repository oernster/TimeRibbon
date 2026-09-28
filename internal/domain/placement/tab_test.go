package placement

import "testing"

// FR-614, its acceptance: a vertical ribbon 196 long flush against the right edge collapses to an
// 8 by 196 band flush against that edge; dragged nearer the left edge, to a band on its left side.
func TestTheTabCoversTheSideNearerItsEdge(t *testing.T) {
	t.Parallel()
	tall := Size{Width: 176, Height: 196}
	flush := Point{X: primary.Work.Right - tall.Width, Y: 400}
	if got, want := Tab(flush, tall, primary.Work, true, Right, TabThickness), (Rect{Left: 1912, Top: 400, Right: 1920, Bottom: 596}); got != want {
		t.Errorf("flush right: got %+v, want %+v", got, want)
	}
	dragged := Point{X: 100, Y: 400}
	if got, want := Tab(dragged, tall, primary.Work, true, Right, TabThickness), (Rect{Left: 100, Top: 400, Right: 108, Bottom: 596}); got != want {
		t.Errorf("nearer the left: got %+v, want %+v", got, want)
	}
	wide := Size{Width: 600, Height: 92}
	if got, want := Tab(Point{X: 700, Y: 0}, wide, primary.Work, false, Top, TabThickness), (Rect{Left: 700, Top: 0, Right: 1300, Bottom: 8}); got != want {
		t.Errorf("horizontal at the top: got %+v, want %+v", got, want)
	}
	low := Point{X: 700, Y: primary.Work.Bottom - wide.Height}
	if got, want := Tab(low, wide, primary.Work, false, Top, TabThickness), (Rect{Left: 700, Top: 1024, Right: 1300, Bottom: 1032}); got != want {
		t.Errorf("horizontal at the bottom: got %+v, want %+v", got, want)
	}
}

// FR-614: at an equal distance from both edges the tab goes to the side of the home edge.
func TestAnEvenDistanceGoesToTheHomeEdge(t *testing.T) {
	t.Parallel()
	tall := Size{Width: 120, Height: 300}
	middle := Point{X: (primary.Work.Width() - tall.Width) / 2, Y: 0}
	if got := Tab(middle, tall, primary.Work, true, Right, TabThickness); got.Right != middle.X+tall.Width {
		t.Errorf("vertical, home right: got %+v, want the right side", got)
	}
	if got := Tab(middle, tall, primary.Work, true, Left, TabThickness); got.Left != middle.X {
		t.Errorf("vertical, home left: got %+v, want the left side", got)
	}
	wide := Size{Width: 600, Height: 92}
	centre := Point{X: 0, Y: (primary.Work.Height() - wide.Height) / 2}
	if got := Tab(centre, wide, primary.Work, false, Top, TabThickness); got.Top != centre.Y {
		t.Errorf("horizontal, home top: got %+v, want the top side", got)
	}
	if got := Tab(centre, wide, primary.Work, false, Bottom, TabThickness); got.Bottom != centre.Y+wide.Height {
		t.Errorf("horizontal, home bottom: got %+v, want the bottom side", got)
	}
}
