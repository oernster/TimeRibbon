//go:build !windows

package desktop

import "github.com/oernster/timeribbon/internal/domain/placement"

// Cursor answers that the pointer cannot be read here, so the corner grip's drag keeps the page's
// own reading of it (FR-623). Reading it on macOS and Linux is open in TECH_DEBT.md.
func Cursor() (placement.Point, bool) { return placement.Point{}, false }
