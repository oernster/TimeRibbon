// Package appdata answers the folder TimeStrip keeps its settings and log in: %APPDATA%\TimeStrip
// (FR-701, NFR-O-1). Nothing is made here; the store and the log make the folder when they write.
package appdata

import (
	"errors"
	"path/filepath"

	"github.com/oernster/timestrip/internal/product"
)

// variable is the environment variable naming the user's roaming application data folder.
const variable = "APPDATA"

// ErrNoAppData is answered when the environment names no application data folder.
var ErrNoAppData = errors.New(variable + " is not set")

// Dir answers the folder, reading the environment through lookup.
func Dir(lookup func(string) (string, bool)) (string, error) {
	base, ok := lookup(variable)
	if !ok || base == "" {
		return "", ErrNoAppData
	}
	return filepath.Join(base, product.Name), nil
}
