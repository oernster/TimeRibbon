//go:build !windows

package desktop

import "github.com/oernster/timeribbon/internal/domain/placement"

// Shape leaves the window a rectangle: on macOS and Linux it is not cut to the ribbon and its map
// (FR-913).
func Shape(Window, []placement.Rect) error { return nil }
