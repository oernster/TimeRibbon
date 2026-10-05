package structural

// TimeRibbon's half of the wire: the Go structs with json tags in dto.go, stated again as
// TypeScript in frontend/src/wire.ts; also the word app.go sends the page. The window's half is
// ribbonkit's and held by the kit's own tests.

import (
	"path/filepath"
	"testing"

	"github.com/oernster/ribbonkit/structure"
)

// wirePairs names each Go wire type with the TypeScript interface stating it again.
var wirePairs = map[string]string{
	"sizeDTO": "Size", "layoutDTO": "Layout", "cellDTO": "Cell", "snapshotDTO": "Snapshot", "placeDTO": "Place",
	"markDTO": "Mark", "sunMapDTO": "SunMap", "textSamplesDTO": "TextSamples", "measuredDTO": "Measured",
}

// appWords names each constant app.go sends the page with the shape the page must state its value
// in, since a view or a test may share the bare word.
var appWords = map[string]string{"openAtAddClock": "const addClock = '%s'"}

func TestTheWireIsStatedAlikeOnBothSides(t *testing.T) {
	root := structure.Root(t)
	structure.CheckTheWireIsStatedAlikeOnBothSides(t, wirePairs,
		[]string{filepath.Join(root, "dto.go")}, []string{filepath.Join(root, "frontend", "src", "wire.ts")})
}

func TestThePageNamesEveryEventAppEmits(t *testing.T) {
	structure.CheckThePageNamesEveryWord(t, filepath.Join(structure.Root(t), "app.go"), appWords, pageFiles(t))
}
