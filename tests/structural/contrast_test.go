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
	"fmt"
	"testing"

	"github.com/oernster/ribbonkit/structure"
)

// minContrast is WCAG 2.x's AA floor for body text (success criterion 1.4.3).
const minContrast = 4.5

// textTokens are the colours the ribbon's words are drawn in; backgrounds are what they lie on.
var (
	textTokens  = []string{"text", "text-muted", "problem"}
	backgrounds = []string{"cell", "surface"}
)

func TestTextMeetsTheContrastFloorOnEverySchemeAndTheme(t *testing.T) {
	halves := []structure.Half{kitHalf(t), dialsHalf(t)}
	for scheme, sides := range structure.Palettes(t, halves, offered()) {
		for side, colourOf := range sides {
			for _, token := range textTokens {
				for _, background := range backgrounds {
					name := fmt.Sprintf("%s %s --%s on --%s", scheme, side, token, background)
					fore, err := colourOf(token)
					if err != nil {
						t.Errorf("%s: %v", name, err)
						continue
					}
					back, err := colourOf(background)
					if err != nil {
						t.Errorf("%s: %v", name, err)
						continue
					}
					ratio, err := structure.Contrast(fore, back)
					if err != nil {
						t.Errorf("%s: %v", name, err)
						continue
					}
					if ratio < minContrast {
						t.Errorf("%s: %s on %s is %.2f:1, under %.1f:1", name, fore, back, ratio, minContrast)
					}
				}
			}
		}
	}
}
