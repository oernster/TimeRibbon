package structural

// Label, time, date and zone mark text meets a contrast ratio of at least 4.5:1 in both themes, on
// every colour scheme the menus offer (NFR-U-1, FR-606, FR-611).
//
// The colours are read from the very files the page loads, ribbonkit's half and TimeRibbon's
// together as the page cascades them; a scheme that leaves a token out draws it in Classic's value
// for the same theme. The label, time and date draw in --text and the zone mark in --text-muted
// (app.css); --problem is checked too, since colours.css states it meets the same floor and the
// colour tests let schemes inherit it on that promise. The cell paints no background of its own, so
// its text lies on #root's --surface; both --cell and --surface are checked. The test lives here
// rather than in Vitest because Vitest hands a CSS import back empty.

import (
	"testing"

	"github.com/oernster/ribbonkit/structure"
)

func TestTextMeetsTheContrastFloorOnEverySchemeAndTheme(t *testing.T) {
	halves := []structure.Half{kitHalf(t), dialsHalf(t)}
	structure.CheckTextMeetsTheContrastFloor(t, halves, structure.Offered(), structure.TextTokens(), structure.Backgrounds())
}
