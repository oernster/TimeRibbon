package structural

// NFR-S-1: the application's one network request is the update check (FR-509), so no Go file of
// this module outside the update package may import a network package. What this cannot see: a
// request Wails or its web view makes on its own account, which is why the rule is held over this
// module's source rather than over the binary.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// networkExempt is the one directory whose files may import a network package: the update check's
// adapter, which asks GitHub for the latest release (Amendment 15).
var networkExempt = filepath.Join("ribbonkit", "infrastructure", "update")

// networkPackages are the packages a request is made through; a path beneath one is one of them.
var networkPackages = []string{"net", "crypto/tls", "golang.org/x/net"}

func isNetworkPackage(imported string) bool {
	for _, network := range networkPackages {
		if imported == network || strings.HasPrefix(imported, network+"/") {
			return true
		}
	}
	return false
}

func TestOnlyTheUpdateCheckImportsANetworkPackage(t *testing.T) {
	exempt := filepath.Join(repoRoot(t), networkExempt)
	for _, path := range goFiles(t) {
		if filepath.Dir(path) == exempt {
			continue
		}
		for _, imported := range importsOf(t, path) {
			if isNetworkPackage(imported) {
				t.Errorf("%s imports %s; only the update check reaches the network (NFR-S-1)", path, imported)
			}
		}
	}
}

// The exemption must name a directory that exists, so a move of the update package cannot leave it
// pointing at nothing while the package's new home goes unchecked.
func TestTheNetworkExemptionNamesTheUpdatePackage(t *testing.T) {
	info, err := os.Stat(filepath.Join(repoRoot(t), networkExempt))
	if err != nil || !info.IsDir() {
		t.Errorf("%s is exempt but is not a directory: %v", filepath.ToSlash(networkExempt), err)
	}
}

func TestNetworkPackageRecognitionIsExact(t *testing.T) {
	for _, network := range []string{"net", "net/http", "crypto/tls", "golang.org/x/net/html"} {
		if !isNetworkPackage(network) {
			t.Errorf("%s was not recognised", network)
		}
	}
	for _, ordinary := range []string{"network-free", "netlify", "crypto/sha256"} {
		if isNetworkPackage(ordinary) {
			t.Errorf("%s was taken for a network package", ordinary)
		}
	}
}
