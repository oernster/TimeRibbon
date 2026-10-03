package structural

// NFR-S-1, beyond the import rule in network_test.go: a request can also leave through a program
// started for it, a system library loaded by name or the page itself. These rules hold each of those
// doors to the ones TimeRibbon uses today. What they cannot see: code reached through cgo's own C,
// a library Wails or the web view loads on its own account and a name built at run time.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// processStarters are the files that may start a program or hand an address to the desktop, each
// for the reason given. Every one hands work to the desktop or to setup; none asks a network.
var processStarters = map[string]string{
	"internal/infrastructure/desktop/browser_unix.go":    "hands an address to open or xdg-open",
	"internal/infrastructure/desktop/browser_windows.go": "hands an address to ShellExecute",
	"internal/infrastructure/setup/deletion.go":          "deletes the install folder after setup exits",
	"internal/infrastructure/setup/process.go":           "starts the installed application",
}

// processImport and processCalls are how a Go file starts another program.
const processImport = "os/exec"

var processCalls = []string{
	"os.StartProcess", "syscall.StartProcess", "syscall.ForkExec", "syscall.CreateProcess",
	"windows.CreateProcess", "windows.ShellExecute",
}

// systemLibraries are the libraries the Windows build loads by name; a request library such as
// winhttp.dll or ws2_32.dll is not among them.
var systemLibraries = []string{"user32.dll", "shell32.dll", "kernel32.dll", "gdi32.dll", "shcore.dll"}

// libraryName matches a string naming a Windows library.
var libraryName = regexp.MustCompile(`(?i)\.dll$`)

// pageRequest matches what makes the page ask a network: a request API or an address to fetch.
// The SVG namespace names a vocabulary, which nothing fetches.
var (
	pageRequest  = regexp.MustCompile(`\bfetch\s*\(|XMLHttpRequest|WebSocket|EventSource|sendBeacon|importScripts|WebTransport|RTCPeerConnection|https?://`)
	svgNamespace = "http://www.w3.org/"
)

// shippedGoFiles is every Go file of the module but its tests, which never ship.
func shippedGoFiles(t *testing.T) []string {
	t.Helper()
	var shipped []string
	for _, path := range goFiles(t) {
		if !strings.HasSuffix(path, "_test.go") {
			shipped = append(shipped, path)
		}
	}
	return shipped
}

// startsProcess answers whether file imports os/exec or calls one of processCalls.
func startsProcess(file *ast.File) bool {
	for _, item := range file.Imports {
		if strings.Trim(item.Path.Value, `"`) == processImport {
			return true
		}
	}
	found := false
	ast.Inspect(file, func(node ast.Node) bool {
		if selector, ok := node.(*ast.SelectorExpr); ok {
			if name, ok := selector.X.(*ast.Ident); ok && slices.Contains(processCalls, name.Name+"."+selector.Sel.Name) {
				found = true
			}
		}
		return !found
	})
	return found
}

// librariesNamed answers every string in file naming a Windows library.
func librariesNamed(file *ast.File) []string {
	var named []string
	ast.Inspect(file, func(node ast.Node) bool {
		if literal, ok := node.(*ast.BasicLit); ok && literal.Kind == token.STRING {
			if text, err := strconv.Unquote(literal.Value); err == nil && libraryName.MatchString(text) {
				named = append(named, text)
			}
		}
		return true
	})
	return named
}

// pageAsks answers each request the page text makes.
func pageAsks(text string) []string {
	return pageRequest.FindAllString(strings.ReplaceAll(text, svgNamespace, ""), -1)
}

func TestOnlyNamedFilesStartAProcess(t *testing.T) {
	root := repoRoot(t)
	for _, path := range shippedGoFiles(t) {
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatalf("parsing %s: %v", path, err)
		}
		relative, _ := filepath.Rel(root, path)
		if _, named := processStarters[filepath.ToSlash(relative)]; startsProcess(file) && !named {
			t.Errorf("%s starts a program; only the files in processStarters may (NFR-S-1)", path)
		}
		for _, library := range librariesNamed(file) {
			if !slices.Contains(systemLibraries, strings.ToLower(library)) {
				t.Errorf("%s names %s, a library not in systemLibraries (NFR-S-1)", path, library)
			}
		}
	}
}

// Each named starter must exist, so a moved file cannot leave its permission behind for another.
func TestEveryNamedProcessStarterExists(t *testing.T) {
	for path := range processStarters {
		if _, err := os.Stat(filepath.Join(repoRoot(t), filepath.FromSlash(path))); err != nil {
			t.Errorf("%s may start a program but is not there: %v", path, err)
		}
	}
}

func TestThePageMakesNoRequest(t *testing.T) {
	pages := append(frontendFiles(t), setupFrontendFiles(t)...)
	pages = append(pages, filepath.Join(repoRoot(t), "frontend", "index.html"))
	for _, path := range pages {
		if strings.Contains(filepath.Base(path), ".test.") {
			continue
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if asked := pageAsks(string(raw)); len(asked) > 0 {
			t.Errorf("%s makes a request (%v); only Go's update check asks a network (NFR-S-1)", path, asked)
		}
	}
}

func TestRequestRecognitionIsExact(t *testing.T) {
	for _, plant := range []string{"fetch('x')", "window.fetch (url)", "new XMLHttpRequest()", "new WebSocket(u)",
		"navigator.sendBeacon(u)", `url("https://x.test/a.woff")`, "http://x.test"} {
		if len(pageAsks(plant)) == 0 {
			t.Errorf("%q was not recognised", plant)
		}
	}
	for _, ordinary := range []string{"prefetch(x)", "api.fetchClocks(x)", "fetched", `xmlns="http://www.w3.org/2000/svg"`} {
		if asked := pageAsks(ordinary); len(asked) > 0 {
			t.Errorf("%q was taken for a request: %v", ordinary, asked)
		}
	}
	source := `package p; import "os/exec"; var _ = exec.Command`
	call := `package p; import "golang.org/x/sys/windows"; func f() { windows.ShellExecute(0, nil, nil, nil, nil, 0) }`
	library := `package p; var _ = "WinHTTP.DLL"`
	for name, text := range map[string]string{"import": source, "call": call} {
		file, err := parser.ParseFile(token.NewFileSet(), name+".go", text, 0)
		if err != nil || !startsProcess(file) {
			t.Errorf("the %s was not recognised as starting a program (%v)", name, err)
		}
	}
	file, err := parser.ParseFile(token.NewFileSet(), "library.go", library, 0)
	if err != nil || !slices.Equal(librariesNamed(file), []string{"WinHTTP.DLL"}) {
		t.Errorf("the library was not recognised (%v)", err)
	}
}
