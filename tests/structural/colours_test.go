package structural

// The menus offer the schemes internal/domain/settings lists, while the page draws them from
// frontend/src/colours.css (FR-611). Nothing else ties the two: a scheme offered with no block draws
// as Classic and a block no menu offers is dead weight, neither of which fails anything at run time.
// A block missing one of Classic's tokens silently borrows Classic's value for it. This test holds
// the list, the blocks and the tokens to one another.

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

func readFrontend(t *testing.T, name string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(repoRoot(t), "frontend", "src", name))
	if err != nil {
		t.Fatalf("reading %s: %v", name, err)
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
	classic := classicBlock.FindStringSubmatch(readFrontend(t, "theme.css"))
	if classic == nil {
		t.Fatal("theme.css has no :root block")
	}
	blocks := map[string]string{}
	for _, match := range schemeBlock.FindAllStringSubmatch(readFrontend(t, "colours.css"), -1) {
		blocks[match[1]] = match[2]
	}
	for _, colour := range ribbon.Colours {
		if colour == ribbon.Classic {
			continue
		}
		body, found := blocks[string(colour)]
		if !found {
			t.Errorf("%s is offered but colours.css has no block for it", colour)
			continue
		}
		stated := tokensIn(body)
		for _, token := range tokensIn(classic[1]) {
			if !slices.Contains(stated, token) && !slices.Contains(optionalTokens, token) {
				t.Errorf("%s leaves --%s to Classic's value", colour, token)
			}
		}
	}
	for name := range blocks {
		if name == string(ribbon.Classic) || !slices.Contains(ribbon.Colours, ribbon.Colour(name)) {
			t.Errorf("colours.css holds a block for %q, which no menu offers", name)
		}
	}
}
