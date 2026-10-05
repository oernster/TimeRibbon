package window

import (
	"math"
	"testing"

	"github.com/oernster/timeribbon/ribbonkit/infrastructure/desktop"
)

// The page's ratio is turned into window pixels with the toolkit's own window scale the desktop
// reports, so a desktop that scales windows itself is not scaled twice.
func TestThePagesRatioIsTakenWithTheToolkitsScale(t *testing.T) {
	t.Parallel()
	app, _, seen, _ := newTestApp(t)
	const ratio, toolkit = 3, 2
	seen.toolkitScale = toolkit
	if err := app.SetPixelRatio(ratio); err != nil {
		t.Fatal(err)
	}
	want := desktop.PixelsPerDIP(ratio, toolkit)
	if got := math.Float64frombits(app.pixelsPerDIP.Load()); got != want {
		t.Errorf("ratio %v at toolkit scale %d sized windows at %v pixels to a unit, want %v", ratio, toolkit, got, want)
	}
}
