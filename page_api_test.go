package main

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"

	"github.com/oernster/timeribbon/ribbonkit/ui/window"
)

// bridgeInterface finds the page's statement of the facade (interface Bridge in api.ts);
// bridgeMethod finds each method named in it.
var (
	bridgeInterface = regexp.MustCompile(`(?s)interface Bridge \{(.*?)\n\}`)
	bridgeMethod    = regexp.MustCompile(`(?m)^\s+(\w+)\(`)
)

// pageCalls answers every method the page calls on the facade, as api.ts states them.
func pageCalls(t *testing.T) []string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("frontend", "src", "api.ts"))
	if err != nil {
		t.Fatal(err)
	}
	found := bridgeInterface.FindSubmatch(raw)
	if found == nil {
		t.Fatal("api.ts states no interface Bridge")
	}
	var names []string
	for _, method := range bridgeMethod.FindAllSubmatch(found[1], -1) {
		names = append(names, string(method[1]))
	}
	return names
}

// Wails binds every exported method of the App it is handed, the window's promoted through the
// embedded Window included; the page reaches each by name (window.go.main.App). Every method the page
// calls must be in that set, whichever half of the facade holds it.
func TestEveryMethodThePageCallsIsBound(t *testing.T) {
	app, _ := newApp(nil, window.Config{Log: io.Discard})
	bound := reflect.TypeOf(app)
	calls := pageCalls(t)
	if len(calls) == 0 {
		t.Fatal("api.ts's Bridge names no method")
	}
	for _, name := range calls {
		if _, ok := bound.MethodByName(name); !ok {
			t.Errorf("the page calls %s, which the bound App does not have", name)
		}
	}
}

// The Control is how TimeRibbon reaches its window; none of it may be bound, so the page can never
// run the window's life, fit it or end it.
func TestNothingOfTheControlIsBound(t *testing.T) {
	bound := reflect.TypeOf(&App{})
	control := reflect.TypeOf(&window.Control{})
	for index := range control.NumMethod() {
		name := control.Method(index).Name
		if _, ok := bound.MethodByName(name); ok {
			t.Errorf("the Control's %s is bound to the page", name)
		}
	}
}
