package structural

// NFR-S-1: the application makes no network request, so no Go file of this module may import a
// network package. What this cannot see: a request Wails or its web view makes on its own account,
// which is why the rule is held over this module's source rather than over the binary.

import (
	"strings"
	"testing"
)

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

func TestTheModuleImportsNoNetworkPackage(t *testing.T) {
	for _, path := range goFiles(t) {
		for _, imported := range importsOf(t, path) {
			if isNetworkPackage(imported) {
				t.Errorf("%s imports %s; the application makes no network request (NFR-S-1)", path, imported)
			}
		}
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
