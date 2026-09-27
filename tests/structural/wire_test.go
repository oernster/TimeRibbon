package structural

// The wire is stated twice: Go structs with json tags in dto.go and TypeScript interfaces in
// frontend/src/wire.ts. The type checker sees only the TypeScript and the marshaller sees only the
// Go, so this test compares them, field for field in both directions. It also holds the event words
// app.go emits to the page, which the page must name exactly.

import (
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
}

// eventWords are the constants in app.go the page must name.
var eventWords = []string{"eventRefresh", "eventOpenSettings", "openAtAddClock"}

var (
	tsInterface = regexp.MustCompile(`(?s)export interface (\w+) \{(.*?)\n\}`)
	tsField     = regexp.MustCompile(`(?m)^\s+(\w+)\??:`)
)

// goWire answers each struct in dto.go with its json names, sorted.
func goWire(t *testing.T) map[string][]string {
	t.Helper()
	parsed, err := parser.ParseFile(token.NewFileSet(), filepath.Join(repoRoot(t), "dto.go"), nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	out := map[string][]string{}
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
	return out
}

// tsWire answers each interface in wire.ts with its field names, sorted.
func tsWire(t *testing.T) map[string][]string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(repoRoot(t), "frontend", "src", "wire.ts"))
	if err != nil {
		t.Fatal(err)
	}
	out := map[string][]string{}
	for _, match := range tsInterface.FindAllStringSubmatch(string(raw), -1) {
		var names []string
		for _, field := range tsField.FindAllStringSubmatch(match[2], -1) {
			names = append(names, field[1])
		}
		slices.Sort(names)
		out[match[1]] = names
	}
	return out
}

func TestTheWireIsStatedAlikeOnBothSides(t *testing.T) {
	goSide, tsSide := goWire(t), tsWire(t)
	if len(goSide) != len(wirePairs) || len(tsSide) != len(wirePairs) {
		t.Errorf("dto.go states %d types and wire.ts %d; the pairs name %d", len(goSide), len(tsSide), len(wirePairs))
	}
	for goName, tsName := range wirePairs {
		if !slices.Equal(goSide[goName], tsSide[tsName]) {
			t.Errorf("%s has %v in Go but %s has %v in TypeScript", goName, goSide[goName], tsName, tsSide[tsName])
		}
	}
}

func TestThePageNamesEveryEventGoEmits(t *testing.T) {
	root := repoRoot(t)
	parsed, err := parser.ParseFile(token.NewFileSet(), filepath.Join(root, "app.go"), nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var page strings.Builder
	for _, file := range frontendFiles(t) {
		raw, _ := os.ReadFile(file)
		page.Write(raw)
	}
	for _, word := range eventWords {
		object := parsed.Scope.Lookup(word)
		if object == nil {
			t.Fatalf("app.go has no constant %s", word)
		}
		literal := object.Decl.(*ast.ValueSpec).Values[0].(*ast.BasicLit).Value
		value, _ := strconv.Unquote(literal)
		if !strings.Contains(page.String(), "'"+value+"'") {
			t.Errorf("app.go emits %q (%s) but the page never names it", value, word)
		}
	}
}
