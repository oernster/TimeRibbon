// Package appdata answers the folder TimeRibbon keeps its settings and log in (FR-701, NFR-O-1):
// %APPDATA%\TimeRibbon on Windows, the XDG configuration folder's TimeRibbon on Linux. Nothing is
// made here; the store and the log make the folder when they write. Each platform supplies where
// the folder's parent is in a file of its own.
package appdata

import (
	"errors"
	"path/filepath"

	"github.com/oernster/timeribbon/internal/product"
)

// ErrNoAppData is answered when the environment names no application data folder.
var ErrNoAppData = errors.New("the environment names no folder for application data")

// Dir answers the folder, reading the environment through lookup.
func Dir(lookup func(string) (string, bool)) (string, error) {
	parent, ok := base(lookup)
	if !ok {
		return "", ErrNoAppData
	}
	return filepath.Join(parent, product.Name), nil
}
