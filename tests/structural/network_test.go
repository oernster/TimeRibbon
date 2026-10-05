package structural

// NFR-S-1: the application's one network request is the update check (FR-509), which is ribbonkit's
// and held there by the same rules. TimeRibbon's own source imports no network package, starts no
// program and its page asks no network.

import (
	"path/filepath"
	"testing"

	"github.com/oernster/ribbonkit/structure"
)

func TestNothingOfTimeRibbonsImportsANetworkPackage(t *testing.T) {
	structure.CheckOnlyTheExemptImportANetworkPackage(t, structure.Root(t), goFiles(t))
}

func TestNothingOfTimeRibbonsStartsAProcess(t *testing.T) {
	structure.CheckOnlyNamedFilesStartAProcess(t, structure.Root(t), goFiles(t), nil)
}

func TestThePageMakesNoRequest(t *testing.T) {
	page := append(pageFiles(t), filepath.Join(structure.Root(t), "frontend", "index.html"))
	structure.CheckThePageMakesNoRequest(t, page)
}
