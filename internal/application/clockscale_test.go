package application

import (
	"errors"
	"testing"

	"github.com/oernster/timeribbon/internal/domain/settings"
	"github.com/oernster/timeribbon/ribbonkit/domain/placement"
)

// doubled is a scale that draws everything twice as large, so every length is exact.
const doubled = 2 * settings.WholeScale

// FR-623: at twice the scale a ribbon is twice as long and twice as thick, whichever way it runs;
// the scale is kept and shown.
func TestAScaledRibbonGrowsInBothDirections(t *testing.T) {
	t.Parallel()
	for _, orientation := range []settings.Orientation{settings.Horizontal, settings.Vertical} {
		initial := clocks(2)
		initial.Orientation = orientation
		r := newRig(t, initial)
		whole, err := r.service.Launch()
		if err != nil {
			t.Fatal(err)
		}
		if err := r.service.SetScale(doubled); err != nil {
			t.Fatal(err)
		}
		got, err := r.service.Launch()
		if err != nil {
			t.Fatal(err)
		}
		if want := (placement.Size{Width: 2 * whole.Size.Width, Height: 2 * whole.Size.Height}); got.Size != want {
			t.Errorf("%s: %+v, want %+v", orientation, got.Size, want)
		}
		if r.store.last(t).Scale != doubled || r.service.Snapshot().Scale != doubled {
			t.Errorf("%s: the scale was not kept and shown", orientation)
		}
	}
}

// FR-623: while the grip is dragged the ribbon is drawn at the preview and nothing is saved; keeping
// a scale ends the preview; a scale outside the bounds is refused either way.
func TestAPreviewIsDrawnButNotKept(t *testing.T) {
	t.Parallel()
	r := newRig(t, clocks(2))
	saves := len(r.store.saved)
	if err := r.service.PreviewScale(settings.MaxScale); err != nil {
		t.Fatal(err)
	}
	if got := r.service.Snapshot(); got.Scale != settings.MaxScale || got.MinScale != settings.MinScale || got.MaxScale != settings.MaxScale {
		t.Errorf("previewing shows %v within %d to %d", got.Scale, got.MinScale, got.MaxScale)
	}
	if len(r.store.saved) != saves {
		t.Error("a preview was saved")
	}
	if err := r.service.SetScale(settings.MinScale); err != nil || r.service.Snapshot().Scale != settings.MinScale {
		t.Errorf("keeping a scale left %v (%v)", r.service.Snapshot().Scale, err)
	}
	for _, outside := range []int{settings.MinScale - 1, settings.MaxScale + 1} {
		if err := r.service.PreviewScale(float64(outside)); !errors.Is(err, ErrUnknownChoice) {
			t.Errorf("previewing %d answered %v", outside, err)
		}
		if err := r.service.SetScale(outside); !errors.Is(err, ErrUnknownChoice) || r.service.Snapshot().Scale != settings.MinScale {
			t.Errorf("keeping %d answered %v", outside, err)
		}
	}
}

// FR-623: a change of scale grows or shrinks the ribbon from its top-left corner, as a window being
// resized does, so the corner the grip is in follows the pointer; previewed or kept, it is never
// centred along its length again.
func TestAChangeOfScaleKeepsTheCorner(t *testing.T) {
	t.Parallel()
	dragged := placement.Point{X: 1500, Y: 40}
	r := newRig(t, draggedTo(2, settings.Vertical, dragged))
	if _, err := r.service.Launch(); err != nil {
		t.Fatal(err)
	}
	steps := []func() error{
		func() error { return r.service.PreviewScale(settings.WholeScale + settings.WholeScale/4) },
		func() error { return r.service.PreviewScale(settings.WholeScale + settings.WholeScale/2) },
		func() error { return r.service.SetScale(doubled) },
	}
	for index, step := range steps {
		if err := step(); err != nil {
			t.Fatal(err)
		}
		got, err := r.service.Rearrange(dragged)
		if err != nil {
			t.Fatal(err)
		}
		if got.At != dragged {
			t.Errorf("step %d: rearranged to %+v, want the corner kept at %+v", index, got.At, dragged)
		}
	}
}

// FR-623: while the grip is dragged the sun map keeps the size it had when the drag began, still
// adjoining the ribbon, so the window's corner holds still; once the scale is kept it is sized again.
func TestTheSunMapIsHeldWhileTheGripIsDragged(t *testing.T) {
	t.Parallel()
	initial := clocks(6)
	initial.Orientation, initial.SunMap, initial.PullOut = settings.Vertical, true, true
	r := newRig(t, initial)
	before, err := r.service.Launch()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.service.PreviewScale(settings.WholeScale + settings.WholeScale/2); err != nil {
		t.Fatal(err)
	}
	during, err := r.service.Rearrange(before.At)
	if err != nil {
		t.Fatal(err)
	}
	if during.Map.Width() != before.Map.Width() || during.Map.Height() != before.Map.Height() || during.Map.Right != during.At.X {
		t.Errorf("during the drag the map is %+v beside a ribbon at %+v; want %+v's size adjoining it", during.Map, during.At, before.Map)
	}
	if err := r.service.SetScale(settings.WholeScale + settings.WholeScale/2); err != nil {
		t.Fatal(err)
	}
	kept, err := r.service.Rearrange(during.At)
	if err != nil {
		t.Fatal(err)
	}
	if kept.Map.Width() == before.Map.Width() {
		t.Errorf("once kept the map is still %d wide; want it sized for the new scale", kept.Map.Width())
	}
}

// FR-104: a change of clocks at the same scale still centres the ribbon along its new length.
func TestAChangeOfClocksAfterAScaleStillRecentres(t *testing.T) {
	t.Parallel()
	dragged := placement.Point{X: 1500, Y: 40}
	r := newRig(t, draggedTo(2, settings.Vertical, dragged))
	if _, err := r.service.Launch(); err != nil {
		t.Fatal(err)
	}
	if err := r.service.SetScale(doubled); err != nil {
		t.Fatal(err)
	}
	if _, err := r.service.Rearrange(dragged); err != nil {
		t.Fatal(err)
	}
	if _, err := r.service.AddClock("Asia/Kolkata"); err != nil {
		t.Fatal(err)
	}
	got, err := r.service.Rearrange(dragged)
	if err != nil {
		t.Fatal(err)
	}
	if want := (primaryMonitor.Work.Height() - got.Size.Height) / 2; got.At.Y != want {
		t.Errorf("rearranged to %+v, want centred top to bottom at %d", got.At, want)
	}
}

// FR-623, FR-106: the scroll bar keeps the thickness the page measured whatever the scale, since it
// is the web engine's own; the cells beside it are scaled.
func TestTheScrollBarIsNotScaled(t *testing.T) {
	t.Parallel()
	const bar = 12
	crowded := clocks(40)
	crowded.Orientation = settings.Horizontal
	crowded.Scale = doubled
	r := newRig(t, crowded)
	if err := r.service.SetScrollbar(bar); err != nil {
		t.Fatal(err)
	}
	got, err := r.service.Launch()
	if err != nil {
		t.Fatal(err)
	}
	want := doubled*(testLayout.Digital.Height+2*testLayout.Padding)/settings.WholeScale + bar
	if !got.Scrolls || got.Size.Height != want {
		t.Errorf("scrolls %v at %d thick, want %d", got.Scrolls, got.Size.Height, want)
	}
}
