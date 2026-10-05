// Package structural enforces the architecture with tests rather than convention (CON-1, CON-2).
//
// Every assertion here has been proved to bite by planting a violation and watching it fail. An
// assertion never seen to fail is not yet a guard.
package structural

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// lineLimit is the module-size cap. dangerBand is five per cent below it: a file that lands between
// them is refactored down to safeLanding rather than left one edit away from breaching.
const (
	lineLimit   = 400
	dangerBand  = lineLimit - lineLimit/20
	safeLanding = 350
)

// modulePath prefixes every internal import.
const modulePath = "github.com/oernster/timeribbon/"

// compositionRoot names the files allowed to import both application and infrastructure: main.go
// builds the adapters; app.go is the facade the window calls; window_life.go is the facade's own
// window handling, apart from app.go only to keep each file small.
var compositionRoot = map[string]bool{"main.go": true, "app.go": true, "window_life.go": true}

// forbiddenInDomain names the packages that would give the domain IO, randomness or a tz database
// of its own. Zones reach it already resolved (CON-5).
var forbiddenInDomain = []string{
	"net", "net/http", "os", "os/exec", "path/filepath", "io/ioutil", "math/rand", "math/rand/v2",
	"time/tzdata", "syscall", "golang.org/x/sys/windows",
}

// forbiddenCallsInDomain read the wall clock or load a zone, which would make the domain's answers
// depend on the machine (FR-207).
var forbiddenCallsInDomain = []string{"time.Now", "time.Since", "time.Until", "time.LoadLocation"}

// skippedDirectories are walked past: other people's code and generated output.
var skippedDirectories = map[string]bool{"frontend": true, ".git": true, "build": true, "node_modules": true}

// repoRoot walks up from the test's directory to the directory holding go.mod.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("no go.mod above the test directory")
		}
		dir = parent
	}
}

// goFiles returns every Go source file in the repository.
func goFiles(t *testing.T) []string {
	t.Helper()
	var found []string
	err := filepath.WalkDir(repoRoot(t), func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && skippedDirectories[entry.Name()] {
			return filepath.SkipDir
		}
		if !entry.IsDir() && strings.HasSuffix(path, ".go") {
			found = append(found, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking the repository: %v", err)
	}
	if len(found) == 0 {
		t.Fatal("no Go files found, the walk is wrong")
	}
	return found
}

// importsOf parses a file and returns its import paths.
func importsOf(t *testing.T, path string) []string {
	t.Helper()
	parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parsing %s: %v", path, err)
	}
	out := make([]string, 0, len(parsed.Imports))
	for _, item := range parsed.Imports {
		out = append(out, strings.Trim(item.Path.Value, `"`))
	}
	return out
}

// kitTree is the directory holding ribbonkit, the desktop behaviour shared with WeatherRibbon. It has
// the same layers as internal and imports nothing outside itself.
const kitTree = "ribbonkit"

// layerOf returns the architectural layer a file belongs to; empty outside internal and the kit.
func layerOf(root, path string) string {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return ""
	}
	parts := strings.Split(filepath.ToSlash(relative), "/")
	if len(parts) >= 2 && (parts[0] == "internal" || parts[0] == kitTree) {
		return parts[1]
	}
	return ""
}

// treeOf returns the top directory of the module path imported names: internal, the kit or another.
func treeOf(imported string) string {
	inner, ok := strings.CutPrefix(imported, modulePath)
	if !ok {
		return ""
	}
	tree, _, _ := strings.Cut(inner, "/")
	return tree
}

// inLayer answers whether imported is a package of layer, in internal or the kit.
func inLayer(imported, layer string) bool {
	return strings.Contains(imported, "internal/"+layer) || strings.Contains(imported, kitTree+"/"+layer)
}

func TestDomainHasNoOutwardImports(t *testing.T) {
	root := repoRoot(t)
	for _, path := range goFiles(t) {
		if layerOf(root, path) != "domain" {
			continue
		}
		for _, imported := range importsOf(t, path) {
			if treeOf(imported) != "" && !inLayer(imported, "domain") {
				t.Errorf("%s imports %s: the domain depends on nothing", path, imported)
			}
		}
	}
}

// TestTheKitImportsNothingOfTimeRibbon holds ribbonkit free of anything clock-specific, so it can
// leave this repository whole (WeatherRibbon CON-10).
func TestTheKitImportsNothingOfTimeRibbon(t *testing.T) {
	root := repoRoot(t)
	for _, path := range goFiles(t) {
		relative, _ := filepath.Rel(root, path)
		if !strings.HasPrefix(filepath.ToSlash(relative), kitTree+"/") {
			continue
		}
		for _, imported := range importsOf(t, path) {
			if tree := treeOf(imported); tree != "" && tree != kitTree {
				t.Errorf("%s imports %s: the kit depends on nothing of TimeRibbon's", relative, imported)
			}
		}
	}
}

func TestDomainIsPure(t *testing.T) {
	root := repoRoot(t)
	for _, path := range goFiles(t) {
		if layerOf(root, path) != "domain" || strings.HasSuffix(path, "_test.go") {
			continue
		}
		for _, imported := range importsOf(t, path) {
			for _, banned := range forbiddenInDomain {
				if imported == banned {
					t.Errorf("%s imports %q: the domain performs no IO", path, banned)
				}
			}
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("reading %s: %v", path, err)
		}
		for _, call := range forbiddenCallsInDomain {
			if strings.Contains(string(raw), call+"(") {
				t.Errorf("%s calls %s: take the instant or the zone as an argument", path, call)
			}
		}
	}
}

func TestApplicationDoesNotImportInfrastructure(t *testing.T) {
	root := repoRoot(t)
	for _, path := range goFiles(t) {
		if layerOf(root, path) != "application" {
			continue
		}
		for _, imported := range importsOf(t, path) {
			if inLayer(imported, "infrastructure") || strings.Contains(imported, "wails") {
				t.Errorf("%s imports %s: the application depends on ports only", path, imported)
			}
		}
	}
}

func TestWailsStaysOutOfInfrastructure(t *testing.T) {
	root := repoRoot(t)
	for _, path := range goFiles(t) {
		if layerOf(root, path) != "infrastructure" {
			continue
		}
		for _, imported := range importsOf(t, path) {
			if strings.Contains(imported, "wails") && !strings.Contains(path, "installer") {
				t.Errorf("%s imports %s: Wails belongs to the composition root", path, imported)
			}
		}
	}
}

func TestCompositionRootIsWhitelisted(t *testing.T) {
	root := repoRoot(t)
	for _, path := range goFiles(t) {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		var application, infrastructure bool
		for _, imported := range importsOf(t, path) {
			application = application || inLayer(imported, "application")
			infrastructure = infrastructure || inLayer(imported, "infrastructure")
		}
		if !application || !infrastructure {
			continue
		}
		relative, _ := filepath.Rel(root, path)
		if !compositionRoot[filepath.ToSlash(relative)] {
			t.Errorf("%s wires application to infrastructure: only the composition root may", relative)
		}
	}
}

// frontendExtensions are the page's source files the size rule governs.
var frontendExtensions = map[string]bool{".ts": true, ".tsx": true, ".css": true}

// frontendFiles returns every source file under frontend/src.
func frontendFiles(t *testing.T) []string {
	t.Helper()
	source := filepath.Join(repoRoot(t), "frontend", "src")
	var found []string
	err := filepath.WalkDir(source, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && frontendExtensions[filepath.Ext(path)] {
			found = append(found, path)
		}
		return nil
	})
	if err != nil || len(found) == 0 {
		t.Fatalf("no front-end source under %s (%v), the walk is wrong", source, err)
	}
	return found
}

// sourceFiles is every file the size rule governs: the Go, the page and the setup program's page.
func sourceFiles(t *testing.T) []string {
	t.Helper()
	files := append(goFiles(t), frontendFiles(t)...)
	return append(files, setupFrontendFiles(t)...)
}

func TestNoFileExceedsLineLimit(t *testing.T) {
	for _, path := range sourceFiles(t) {
		if count := lineCount(t, path); count > lineLimit {
			t.Errorf("%s has %d lines, over the %d limit", path, count, lineLimit)
		}
	}
}

func TestNoFileInDangerBand(t *testing.T) {
	for _, path := range sourceFiles(t) {
		count := lineCount(t, path)
		if count > dangerBand && count <= lineLimit {
			t.Errorf("%s has %d lines, inside the danger band %d to %d: reduce it to %d or fewer",
				path, count, dangerBand+1, lineLimit, safeLanding)
		}
	}
}

// lineCount counts the lines in a file as an editor numbers them: a newline ends a line rather
// than starting one.
func lineCount(t *testing.T, path string) int {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	text := string(raw)
	lines := strings.Count(text, "\n")
	if text != "" && !strings.HasSuffix(text, "\n") {
		lines++
	}
	return lines
}

func TestEveryExportedTypeIsDocumented(t *testing.T) {
	for _, path := range goFiles(t) {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("parsing %s: %v", path, err)
		}
		for _, declaration := range parsed.Decls {
			general, ok := declaration.(*ast.GenDecl)
			if !ok || general.Tok != token.TYPE || general.Doc != nil {
				continue
			}
			for _, spec := range general.Specs {
				typed, ok := spec.(*ast.TypeSpec)
				if ok && typed.Name.IsExported() && typed.Doc == nil {
					t.Errorf("%s: exported type %s has no doc comment", path, typed.Name.Name)
				}
			}
		}
	}
}
