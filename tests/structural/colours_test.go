package structural

// The menus offer the schemes ribbonkit's ribbon domain lists, while the page draws them from two
// halves of the palette: ribbonkit's (the ribbon's own tokens, held by the kit's own tests) and
// TimeRibbon's dials.css (its content's), held here. A scheme offered with no block draws as
// Classic; a block no menu offers is dead weight; a block missing one of Classic's tokens silently
// borrows Classic's value (FR-611).

import (
	"path/filepath"
	"testing"

	"github.com/oernster/ribbonkit/domain/ribbon"
	"github.com/oernster/ribbonkit/structure"
)

// optionalTokens may be left to Classic's value: its problem colour already meets the contrast floor
// on every scheme's cell and surface (NFR-U-1).
var optionalTokens = []string{"problem"}

// dialsHalf is TimeRibbon's half of the palette: dials.css states Classic and every other scheme.
func dialsHalf(t *testing.T) structure.Half {
	t.Helper()
	dials := filepath.Join(structure.Root(t), "frontend", "src", "dials.css")
	return structure.Half{Classic: dials, Schemes: dials}
}

// kitHalf is ribbonkit's half, read from the kit Go builds against: theme.css and colours.css.
func kitHalf(t *testing.T) structure.Half {
	t.Helper()
	web := filepath.Join(kitDir(t), "web")
	return structure.Half{Classic: filepath.Join(web, "theme.css"), Schemes: filepath.Join(web, "colours.css")}
}

// offered answers the schemes the menus offer, by name.
func offered() []string {
	names := make([]string, 0, len(ribbon.Colours))
	for _, colour := range ribbon.Colours {
		names = append(names, string(colour))
	}
	return names
}

func TestEveryOfferedSchemeHasItsOwnCompleteBlock(t *testing.T) {
	structure.CheckEveryOfferedSchemeHasItsOwnCompleteBlock(t, dialsHalf(t), string(ribbon.Classic), offered(), optionalTokens)
}

// The page shows the system's dark under System and the chosen dark under Dark (FR-606).
func TestClassicDarkIsTheSameUnderTheSystemAsWhenChosen(t *testing.T) {
	structure.CheckClassicDarkIsTheSameUnderTheSystemAsWhenChosen(t, dialsHalf(t))
}
