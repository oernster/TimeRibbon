package structural

// The kit's layers are held by the boundary tests, which know a file's layer by the folder under
// ribbonkit it sits in. Two folders are not layers: web (the page's half) plus installer (the setup
// program's window over the install policy, a program's composition rather than a layer). These
// tests hold the kit to those folders alone, so a new one cannot escape the layer rules unnoticed.
// They also hold the setup program to the one package of the module it is the window over.

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// kitFolders are the folders the kit may hold: its four layers, the page's half and the setup
// program.
var kitFolders = []string{"application", "domain", "infrastructure", "ui", "web", "installer"}

// kitSetupProgram is the setup program's folder; kitInstallPolicy is the one package of the module
// it may import.
var (
	kitSetupProgram  = filepath.Join(kitTree, "installer")
	kitInstallPolicy = modulePath + kitTree + "/infrastructure/setup"
)

func TestTheKitHoldsOnlyItsLayersThePageAndTheSetupProgram(t *testing.T) {
	entries, err := os.ReadDir(filepath.Join(repoRoot(t), kitTree))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() && !slices.Contains(kitFolders, entry.Name()) {
			t.Errorf("%s/%s is neither a layer nor one of the kit's two programs: the layer rules cannot see it", kitTree, entry.Name())
		}
	}
}

func TestTheSetupProgramReachesOnlyTheInstallPolicy(t *testing.T) {
	root := repoRoot(t)
	read := 0
	for _, path := range goFiles(t) {
		relative, _ := filepath.Rel(root, path)
		if !strings.HasPrefix(relative, kitSetupProgram+string(filepath.Separator)) {
			continue
		}
		read++
		for _, imported := range importsOf(t, path) {
			if strings.HasPrefix(imported, modulePath) && imported != kitInstallPolicy {
				t.Errorf("%s imports %s: the setup program is the window over the install policy alone", filepath.ToSlash(relative), imported)
			}
		}
	}
	if read == 0 {
		t.Fatalf("no Go file found under %s, the walk is wrong", filepath.ToSlash(kitSetupProgram))
	}
}
