package structural

// FR-607: About credits every component the application ships. The Go modules are the part a
// machine can check, so this asks the Go tool which modules the application and the setup program
// link as wails build builds them, then holds the credits to that list in both directions. A module
// linked but not credited fails; so does one credited that nothing links any more.

import (
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"

	"github.com/oernster/timestrip/internal/product"
)

// wailsBuildTags are the tags wails build compiles a production build with, which change what the
// Wails module links.
const wailsBuildTags = "desktop,production"

// shippedPackages are the two executables' main packages.
var shippedPackages = []string{".", "./installer"}

// linkedModules answers every module the shipped executables link, this one left out.
func linkedModules(t *testing.T) []string {
	t.Helper()
	args := append([]string{"list", "-tags", wailsBuildTags, "-deps", "-f", "{{with .Module}}{{.Path}}{{end}}"}, shippedPackages...)
	command := exec.Command("go", args...)
	command.Dir = repoRoot(t)
	command.Env = append(os.Environ(), "GOOS=windows", "CGO_ENABLED=0")
	out, err := command.Output()
	if err != nil {
		t.Fatalf("go list: %v", err)
	}
	var modules []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		module := strings.TrimSpace(line)
		if module != "" && module+"/" != modulePath && !slices.Contains(modules, module) {
			modules = append(modules, module)
		}
	}
	if len(modules) == 0 {
		t.Fatal("go list named no module at all, the query is wrong")
	}
	return modules
}

func TestEveryLinkedModuleIsCredited(t *testing.T) {
	var credited []string
	for _, credit := range product.Credits {
		if credit.Module != "" {
			credited = append(credited, credit.Module)
		}
	}
	linked := linkedModules(t)
	for _, module := range linked {
		if !slices.Contains(credited, module) {
			t.Errorf("%s is linked into what ships but About does not credit it", module)
		}
	}
	for _, module := range credited {
		if !slices.Contains(linked, module) {
			t.Errorf("About credits %s, which nothing that ships links any more", module)
		}
	}
}
