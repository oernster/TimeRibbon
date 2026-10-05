package structural

// The wire is stated twice: Go structs with json tags in wireFiles and TypeScript interfaces in
// frontend/src/wire.ts. The type checker sees only the TypeScript and the marshaller sees only the
// Go, so this test compares them, field for field in both directions. It also holds the event words
// the window, app.go and the setup facade emit to their pages, which each page must name exactly.

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// wirePairs names each Go wire type with the TypeScript interface stating it again.
var wirePairs = map[string]string{
	"sizeDTO": "Size", "layoutDTO": "Layout", "cellDTO": "Cell", "snapshotDTO": "Snapshot", "placeDTO": "Place",
	"aboutDTO": "AboutFacts", "creditDTO": "Credit", "updateDTO": "UpdateStatus",
	"Box": "Box", "markDTO": "Mark", "sunMapDTO": "SunMap",
	"textSamplesDTO": "TextSamples", "measuredDTO": "Measured", "choiceDTO": "MenuChoice",
}

// wireFiles are the Go halves of the wire: TimeRibbon's own and the kit window's.
var wireFiles = []string{"dto.go", filepath.Join("ribbonkit", "ui", "window", "wire.go")}

// windowFile is where the kit's window states the words it sends the page.
var windowFile = filepath.Join("ribbonkit", "ui", "window", "facade.go")

// setupFacade is where the kit's setup program states the words it sends its page.
var setupFacade = filepath.Join("ribbonkit", "installer", "facade.go")

// windowWords names each constant the window sends the page with the shape the page must state its
// value in: a listener for an event and a key of panelFor for a panel. appWords does the same for
// app.go, setupWords for the setup facade and the setup page. The shape rather than the bare quoted
// word, since a view, a panel or a test may share the word and would hide a listener or a key that
// no longer matches.
var (
	windowWords = map[string]string{
		"eventRefresh":   heardByTheWindow,
		"eventOpenPanel": heardByTheWindow,
		"openAtSettings": panelKey,
		"openAtAbout":    panelKey,
		"openAtLicence":  panelKey,
		"openAtUpdate":   panelKey,
	}
	appWords   = map[string]string{"openAtAddClock": "const addClock = '%s'"}
	setupWords = map[string]string{"progressEvent": "EventsOn('%s'"}
)

const (
	heardByTheWindow = "on('%s'"
	panelKey         = "'%s': "
)

var (
	tsInterface = regexp.MustCompile(`(?s)export interface (\w+) \{(.*?)\n\}`)
	tsField     = regexp.MustCompile(`(?m)^\s+(\w+)\??:`)
)

// goWire answers each struct in wireFiles with its json names, sorted.
func goWire(t *testing.T) map[string][]string {
	t.Helper()
	out := map[string][]string{}
	for _, file := range wireFiles {
		goWireOf(t, file, out)
	}
	return out
}

// goWireOf adds each struct in file to out with its json names, sorted.
func goWireOf(t *testing.T, file string, out map[string][]string) {
	t.Helper()
	parsed, err := parser.ParseFile(token.NewFileSet(), filepath.Join(repoRoot(t), file), nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	ast.Inspect(parsed, func(node ast.Node) bool {
		spec, ok := node.(*ast.TypeSpec)
		if !ok {
			return true
		}
		fields, ok := spec.Type.(*ast.StructType)
		if !ok {
			return true
		}
		var names []string
		for _, field := range fields.Fields.List {
			tag, _ := strconv.Unquote(field.Tag.Value)
			name := strings.Split(reflect.StructTag(tag).Get("json"), ",")[0]
			names = append(names, name)
		}
		slices.Sort(names)
		out[spec.Name.Name] = names
		return true
	})
}

// tsWire answers each interface in the page's wire.ts files, TimeRibbon's and the kit's, with its
// field names, sorted.
func tsWire(t *testing.T) map[string][]string {
	t.Helper()
	out := map[string][]string{}
	for _, source := range pageSources(t) {
		raw, err := os.ReadFile(filepath.Join(source, "wire.ts"))
		if err != nil {
			t.Fatal(err)
		}
		for _, match := range tsInterface.FindAllStringSubmatch(string(raw), -1) {
			var names []string
			for _, field := range tsField.FindAllStringSubmatch(match[2], -1) {
				names = append(names, field[1])
			}
			slices.Sort(names)
			out[match[1]] = names
		}
	}
	return out
}

func TestTheWireIsStatedAlikeOnBothSides(t *testing.T) {
	goSide, tsSide := goWire(t), tsWire(t)
	if len(goSide) != len(wirePairs) || len(tsSide) != len(wirePairs) {
		t.Errorf("the Go wire states %d types and wire.ts %d; the pairs name %d", len(goSide), len(tsSide), len(wirePairs))
	}
	for goName, tsName := range wirePairs {
		if !slices.Equal(goSide[goName], tsSide[tsName]) {
			t.Errorf("%s has %v in Go but %s has %v in TypeScript", goName, goSide[goName], tsName, tsSide[tsName])
		}
	}
}

func TestThePageNamesEveryEventGoEmits(t *testing.T) {
	requirePageNamesEveryWord(t, windowFile, windowWords, frontendFiles(t))
	requirePageNamesEveryWord(t, "app.go", appWords, frontendFiles(t))
}

// The setup program's progress bar moves only on the word the kit's setup facade emits.
func TestTheSetupPageNamesEveryEventSetupEmits(t *testing.T) {
	requirePageNamesEveryWord(t, setupFacade, setupWords, setupFrontendFiles(t))
}

// requirePageNamesEveryWord fails for each string constant in goFile whose value no file of the
// page states in the shape given for it, since a word one side alone renames reaches nothing and
// nothing fails.
func requirePageNamesEveryWord(t *testing.T, goFile string, words map[string]string, pageFiles []string) {
	t.Helper()
	parsed, err := parser.ParseFile(token.NewFileSet(), filepath.Join(repoRoot(t), goFile), nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var page strings.Builder
	for _, file := range pageFiles {
		raw, _ := os.ReadFile(file)
		page.Write(raw)
	}
	for word, shape := range words {
		object := parsed.Scope.Lookup(word)
		if object == nil {
			t.Fatalf("%s has no constant %s", filepath.ToSlash(goFile), word)
		}
		literal := object.Decl.(*ast.ValueSpec).Values[0].(*ast.BasicLit).Value
		value, _ := strconv.Unquote(literal)
		if !strings.Contains(page.String(), fmt.Sprintf(shape, value)) {
			t.Errorf("%s emits %q (%s) but the page never names it", filepath.ToSlash(goFile), value, word)
		}
	}
}
