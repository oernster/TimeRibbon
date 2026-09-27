package product

// Author is who wrote TimeStrip, as About names them (FR-607).
const Author = "Oliver Ernster"

// Copyright is the copyright line About shows beneath the author (FR-607).
const Copyright = "© " + Author

// Credit is one component the application ships: the Go module it comes from where it is one,
// the name it is credited under, its licence and what it does here.
type Credit struct {
	Module  string
	Name    string
	Licence string
	Role    string
}

// Credits is every component the application ships (FR-607), kept by hand so each names the right
// people in words a reader follows. TestEveryLinkedModuleIsCredited holds the modules here to the
// ones the application and the setup program link, in both directions.
var Credits = []Credit{
	{"", "Go standard library", "BSD-3-Clause", "the language and its runtime"},
	{"", "tz database", "public domain, IANA", "the list of places plus the time zone rules, built in through Go's time/tzdata"},
	{"github.com/wailsapp/wails/v2", "Wails v2", "MIT", "the desktop shell"},
	{"", "React and React DOM", "MIT", "the user interface"},
	{"golang.org/x/sys", "golang.org/x/sys", "BSD-3-Clause", "the tray icon, the displays, the window and Start with Windows"},
	{"github.com/go-ole/go-ole", "go-ole", "MIT", "the setup program's Start Menu and Desktop shortcuts"},
	// Linked in by the modules above rather than named by this application; each role says which.
	{"github.com/wailsapp/go-webview2", "go-webview2", "MIT", "used by Wails"},
	{"github.com/wailsapp/mimetype", "mimetype", "MIT", "used by Wails"},
	{"git.sr.ht/~jackmordaunt/go-toast/v2", "go-toast", "MIT or Unlicense", "used by Wails"},
	{"github.com/bep/debounce", "debounce", "MIT", "used by Wails"},
	{"github.com/google/uuid", "google/uuid", "BSD-3-Clause", "used by Wails"},
	{"github.com/leaanthony/go-ansi-parser", "go-ansi-parser", "MIT", "used by Wails"},
	{"github.com/leaanthony/slicer", "slicer", "MIT", "used by Wails"},
	{"github.com/leaanthony/u", "leaanthony/u", "MIT", "used by Wails"},
	{"github.com/pkg/browser", "pkg/browser", "BSD-2-Clause", "used by Wails"},
	{"github.com/samber/lo", "lo", "MIT", "used by Wails"},
	{"github.com/tkrajina/go-reflector", "go-reflector", "Apache-2.0", "used by Wails"},
	{"golang.org/x/net", "golang.org/x/net", "BSD-3-Clause", "used by Wails and mimetype"},
	{"github.com/pkg/errors", "pkg/errors", "BSD-2-Clause", "used by Wails"},
	{"github.com/rivo/uniseg", "uniseg", "MIT", "used by go-ansi-parser"},
	{"golang.org/x/text", "golang.org/x/text", "BSD-3-Clause", "used by lo"},
}
