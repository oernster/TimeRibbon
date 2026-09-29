package main

import (
	"math"
	"slices"
	"testing"
)

// The page's background reaches the window, so a window catching up with a new size shows it rather
// than white (measured 2026-09-29).
func TestThePagesBackgroundReachesTheWindow(t *testing.T) {
	t.Parallel()
	app, _, seen, _ := newTestApp(t)
	if err := app.SetBackground(7, 36, math.MaxUint8); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(seen.backgrounds, [][3]uint8{{7, 36, math.MaxUint8}}) {
		t.Errorf("painted %v", seen.backgrounds)
	}
}

// The page is foreign input: a channel outside a byte is refused, never wrapped into another colour.
func TestABackgroundThatIsNotAColourIsRefused(t *testing.T) {
	t.Parallel()
	app, _, seen, _ := newTestApp(t)
	for _, channel := range []int{-1, math.MaxUint8 + 1} {
		if err := app.SetBackground(0, channel, 0); err == nil {
			t.Errorf("channel %d was taken", channel)
		}
	}
	if len(seen.backgrounds) != 0 {
		t.Errorf("painted %v", seen.backgrounds)
	}
}
