package application

import (
	"errors"
	"testing"

	"github.com/oernster/ribbonkit/domain/ribbon"
	"github.com/oernster/timeribbon/internal/domain/settings"
)

// FR-622: an opacity within the bounds is chosen, saved and shown with the least allowed; one
// outside them is refused and changes nothing.
func TestOpacityIsChosenSavedAndShown(t *testing.T) {
	t.Parallel()
	r := newRig(t, settings.Defaults())
	if got := r.service.Snapshot(); got.Opacity != ribbon.MaxOpacity || got.MinOpacity != ribbon.MinOpacity {
		t.Errorf("a first run shows %d percent, least %d", got.Opacity, got.MinOpacity)
	}
	chosen := ribbon.MinOpacity
	if err := r.service.SetOpacity(chosen); err != nil || r.store.last(t).Opacity != chosen || r.service.Snapshot().Opacity != chosen {
		t.Errorf("%d percent was not chosen, saved and shown: %v", chosen, err)
	}
	for _, outside := range []int{ribbon.MinOpacity - 1, ribbon.MaxOpacity + 1} {
		if err := r.service.SetOpacity(outside); !errors.Is(err, ribbon.ErrUnknownChoice) || r.service.Snapshot().Opacity != chosen {
			t.Errorf("%d percent answered %v and left %d", outside, err, r.service.Snapshot().Opacity)
		}
	}
}
