package arranger

import (
	"fmt"

	"github.com/oernster/timeribbon/ribbonkit/domain/placement"
	"github.com/oernster/timeribbon/ribbonkit/domain/ribbon"
)

// SetScale chooses how large the cells are drawn on top of their size in percent, then keeps it;
// any preview ends and the map is no longer held (FR-623). A value outside ribbon.MinScale to
// ribbon.MaxScale is refused and changes nothing.
func (a *Arranger) SetScale(percent int) error {
	err := scaleOffered(float64(percent))
	if err == nil {
		err = a.host.ChangeRibbon(func(current ribbon.Choices) ribbon.Choices {
			current.Scale = percent
			return current
		})
	}
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.previewScale, a.held = 0, heldMap{}
	return err
}

// PreviewScale draws the ribbon at percent without keeping it, while its grip is dragged; SetScale
// keeps the scale the drag ends at (FR-623). The percent need not be whole, so the ribbon follows
// the pointer pixel by pixel. The first preview of a drag holds the map where it stands. A value
// outside the bounds is refused.
func (a *Arranger) PreviewScale(percent float64) error {
	if err := scaleOffered(percent); err != nil {
		return err
	}
	a.mutex.Lock()
	defer a.mutex.Unlock()
	if a.previewScale == 0 && a.last.known && a.last.sunMap != (placement.Rect{}) {
		ribbon := placement.Rect{Left: a.last.at.X, Top: a.last.at.Y, Right: a.last.at.X + a.last.size.Width, Bottom: a.last.at.Y + a.last.size.Height}
		a.held = heldMap{known: true, ribbon: ribbon, sunMap: a.last.sunMap}
	}
	a.previewScale = percent
	return nil
}

// DrawnScale answers the scale the ribbon is drawn at while kept is the scale chosen: the preview
// while a drag of the grip is under way, else kept (FR-623).
func (a *Arranger) DrawnScale(kept int) float64 {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	return a.scaleOf(kept)
}

// scaleOf is DrawnScale for a caller holding the mutex.
func (a *Arranger) scaleOf(kept int) float64 {
	if a.previewScale != 0 {
		return a.previewScale
	}
	return float64(kept)
}

// scaleOffered refuses a percent outside ribbon.MinScale to ribbon.MaxScale, which normalising the
// choices would replace.
func scaleOffered(percent float64) error {
	if !(percent >= ribbon.MinScale && percent <= ribbon.MaxScale) {
		return fmt.Errorf("%w: a scale of %v percent", ribbon.ErrUnknownChoice, percent)
	}
	return nil
}

// heldOf answers the map held while the grip is dragged, read under the mutex.
func (a *Arranger) heldOf() heldMap {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	return a.held
}

// read answers the host's ribbon choices, normalised, with its content.
func (a *Arranger) read() (ribbon.Choices, Content) {
	current, content := a.host.Ribbon()
	return current.Normalised(), content
}
