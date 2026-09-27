package structural

// The setup program's page (FR-801 to FR-810) has no build step, so its files are the source
// rather than an output; nothing that walks frontend/src reaches them. They are read here. They are
// held to the size rule (CON-2), the page must load every script beside it and none of them may
// write the product's name, which the page is handed by the setup program.

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/oernster/timestrip/internal/product"
)

// setupFrontendDir is the setup program's page.
var setupFrontendDir = filepath.Join("installer", "frontend", "dist")

// setupFrontendExtensions are the files there the rules govern; the images beside them are artwork.
var setupFrontendExtensions = map[string]bool{".html": true, ".css": true, ".js": true}

// setupPage is the page itself; setupScriptTag matches one script it loads.
var (
	setupPage      = "index.html"
	setupScriptTag = regexp.MustCompile(`<script src="([^"]+)"></script>`)
)

// setupFrontendFiles returns the setup page's own source files.
func setupFrontendFiles(t *testing.T) []string {
	t.Helper()
	dir := filepath.Join(repoRoot(t), setupFrontendDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading %s: %v", filepath.ToSlash(setupFrontendDir), err)
	}
	var found []string
	for _, entry := range entries {
		if !entry.IsDir() && setupFrontendExtensions[strings.ToLower(filepath.Ext(entry.Name()))] {
			found = append(found, filepath.Join(dir, entry.Name()))
		}
	}
	if len(found) == 0 {
		t.Fatalf("no source found in %s, the walk is wrong", filepath.ToSlash(setupFrontendDir))
	}
	return found
}

// A script the page never loads defines nothing that runs, which no test of the script alone can
// notice.
func TestTheSetupPageLoadsEveryScript(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(repoRoot(t), setupFrontendDir, setupPage))
	if err != nil {
		t.Fatalf("reading the setup page: %v", err)
	}
	var loaded []string
	for _, match := range setupScriptTag.FindAllSubmatch(raw, -1) {
		loaded = append(loaded, string(match[1]))
	}
	for _, path := range setupFrontendFiles(t) {
		name := filepath.Base(path)
		if strings.EqualFold(filepath.Ext(name), ".js") && !slices.Contains(loaded, name) {
			t.Errorf("the setup page never loads %s, so nothing it defines runs", name)
		}
	}
}

// The setup page is handed the product's name with the reading of the machine; written down in the
// page, a rename would leave setup announcing a product that no longer exists.
func TestTheSetupPageNeverWritesTheProductsName(t *testing.T) {
	for _, path := range setupFrontendFiles(t) {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("reading %s: %v", path, err)
		}
		if strings.Contains(strings.ToLower(string(raw)), strings.ToLower(product.Name)) {
			t.Errorf("%s writes %s; the page must read the name from the setup program", filepath.Base(path), product.Name)
		}
	}
}
