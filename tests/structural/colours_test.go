package structural

// The menus offer the schemes ribbonkit's ribbon domain lists, while the page draws them from two
// halves of the palette: ribbonkit's (the ribbon's own tokens) and TimeRibbon's dials.css (its
// content's). Nothing else ties them: a scheme offered with no block draws as Classic and a block no
// menu offers is dead weight, neither of which fails anything at run time. A block missing one of
// Classic's tokens in its half silently borrows Classic's value for it. This test holds the list, the
// blocks and the tokens to one another, half by half (FR-611).

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"testing"

	"github.com/oernster/timeribbon/ribbonkit/domain/ribbon"
)

var (
	schemeBlock  = regexp.MustCompile(`(?m)^:root\[data-colour='(\w+)'\] \{([^}]*)\}`)
	classicBlock = regexp.MustCompile(`(?m)^:root \{([^}]*)\}`)
	tokenName    = regexp.MustCompile(`--([\w-]+):`)
)

// optionalTokens may be left to Classic's value: its problem colour already meets the contrast floor
// on every scheme's cell and surface (NFR-U-1).
var optionalTokens = []string{"problem"}

// paletteHalf is one half of the palette, each file relative to the repository: classic states
// Classic in light, under the system's dark and in chosen dark; schemes states every other scheme.
type paletteHalf struct {
	classic string
	schemes string
}

// paletteHalves are ribbonkit's half (the ribbon's own tokens) and TimeRibbon's (its content's).
var paletteHalves = []paletteHalf{
	{classic: filepath.Join(kitTree, "web", "theme.css"), schemes: filepath.Join(kitTree, "web", "colours.css")},
	{classic: filepath.Join("frontend", "src", "dials.css"), schemes: filepath.Join("frontend", "src", "dials.css")},
}

// readPage answers the page file at relative, from the repository's root.
func readPage(t *testing.T, relative string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(repoRoot(t), relative))
	if err != nil {
		t.Fatalf("reading %s: %v", relative, err)
	}
	return string(raw)
}

func tokensIn(body string) []string {
	var names []string
	for _, match := range tokenName.FindAllStringSubmatch(body, -1) {
		names = append(names, match[1])
	}
	return names
}

func TestEveryOfferedSchemeHasItsOwnCompleteBlock(t *testing.T) {
	for _, half := range paletteHalves {
		requireCompleteSchemes(t, half)
	}
}

// requireCompleteSchemes fails for each offered scheme half has no block for, each of Classic's
// tokens a block leaves to Classic and each block no menu offers.
func requireCompleteSchemes(t *testing.T, half paletteHalf) {
	t.Helper()
	classic := classicBlock.FindStringSubmatch(readPage(t, half.classic))
	if classic == nil {
		t.Fatalf("%s has no :root block", half.classic)
	}
	blocks := map[string]string{}
	for _, match := range schemeBlock.FindAllStringSubmatch(readPage(t, half.schemes), -1) {
		blocks[match[1]] = match[2]
	}
	for _, colour := range ribbon.Colours {
		if colour == ribbon.Classic {
			continue
		}
		body, found := blocks[string(colour)]
		if !found {
			t.Errorf("%s is offered but %s has no block for it", colour, half.schemes)
			continue
		}
		stated := tokensIn(body)
		for _, token := range tokensIn(classic[1]) {
			if !slices.Contains(stated, token) && !slices.Contains(optionalTokens, token) {
				t.Errorf("%s leaves --%s to Classic's value in %s", colour, token, half.schemes)
			}
		}
	}
	for name := range blocks {
		if name == string(ribbon.Classic) || !slices.Contains(ribbon.Colours, ribbon.Colour(name)) {
			t.Errorf("%s holds a block for %q, which no menu offers", half.schemes, name)
		}
	}
}
