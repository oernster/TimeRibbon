package application

import (
	"errors"
	"testing"

	"github.com/oernster/timeribbon/internal/domain/settings"
)

// FR-622: an opacity within the bounds is chosen, saved and shown with the least allowed; one
// outside them is refused and changes nothing.
func TestOpacityIsChosenSavedAndShown(t *testing.T) {
	t.Parallel()
	r := newRig(t, settings.Defaults())
	if got := r.service.Snapshot(); got.Opacity != settings.MaxOpacity || got.MinOpacity != settings.MinOpacity {
		t.Errorf("a first run shows %d percent, least %d", got.Opacity, got.MinOpacity)
	}
	chosen := settings.MinOpacity
	if err := r.service.SetOpacity(chosen); err != nil || r.store.last(t).Opacity != chosen || r.service.Snapshot().Opacity != chosen {
		t.Errorf("%d percent was not chosen, saved and shown: %v", chosen, err)
	}
	for _, outside := range []int{settings.MinOpacity - 1, settings.MaxOpacity + 1} {
		if err := r.service.SetOpacity(outside); !errors.Is(err, ErrUnknownChoice) || r.service.Snapshot().Opacity != chosen {
			t.Errorf("%d percent answered %v and left %d", outside, err, r.service.Snapshot().Opacity)
		}
	}
}
