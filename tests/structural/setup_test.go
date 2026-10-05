package structural

// The setup program's page (FR-801 to FR-810) is ribbonkit's and is handed the product's name with
// the reading of the machine. Only TimeRibbon knows its name, so this holds the page it ships to
// never spelling it: written down there, a rename would leave setup announcing a product that no
// longer exists. The page's other rules are the kit's own.

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/oernster/ribbonkit/structure"
	"github.com/oernster/timeribbon/internal/product"
)

// setupPageExtensions are the setup page's own source files; the images beside them are artwork.
var setupPageExtensions = []string{".html", ".css", ".js"}

func TestTheSetupPageNeverWritesTheProductsName(t *testing.T) {
	page := filepath.Join(kitDir(t), "installer", "page")
	for _, path := range structure.FilesWith(t, setupPageExtensions, page) {
		if strings.Contains(strings.ToLower(structure.Read(t, path)), strings.ToLower(product.Name)) {
			t.Errorf("%s writes %s; the page must read the name from the setup program", filepath.Base(path), product.Name)
		}
	}
}
