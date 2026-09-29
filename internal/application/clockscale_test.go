package application

import (
	"errors"
	"testing"

	"github.com/oernster/timeribbon/internal/domain/placement"
	"github.com/oernster/timeribbon/internal/domain/settings"
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
		t.Errorf("previewing shows %d within %d to %d", got.Scale, got.MinScale, got.MaxScale)
	}
	if len(r.store.saved) != saves {
		t.Error("a preview was saved")
	}
	if err := r.service.SetScale(settings.MinScale); err != nil || r.service.Snapshot().Scale != settings.MinScale {
		t.Errorf("keeping a scale left %d (%v)", r.service.Snapshot().Scale, err)
	}
	for _, outside := range []int{settings.MinScale - 1, settings.MaxScale + 1} {
		if err := r.service.PreviewScale(outside); !errors.Is(err, ErrUnknownChoice) {
			t.Errorf("previewing %d answered %v", outside, err)
		}
		if err := r.service.SetScale(outside); !errors.Is(err, ErrUnknownChoice) || r.service.Snapshot().Scale != settings.MinScale {
			t.Errorf("keeping %d answered %v", outside, err)
		}
	}
}

// FR-623, FR-104: a change of scale changes the ribbon's length, so it is centred along it again,
// its side kept, as a change of clocks centres it.
func TestAChangeOfScaleRecentresTheRibbon(t *testing.T) {
	t.Parallel()
	dragged := placement.Point{X: 1700, Y: 40}
	r := newRig(t, draggedTo(2, settings.Vertical, dragged))
	if _, err := r.service.Launch(); err != nil {
		t.Fatal(err)
	}
	if err := r.service.SetScale(doubled); err != nil {
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
