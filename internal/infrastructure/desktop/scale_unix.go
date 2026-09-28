//go:build linux || darwin

package desktop

// unitsPerCSSPixel is how many of the toolkit's units the page's CSS pixel takes: GTK and AppKit
// size the window in DIP and scale the page with it, so one.
const unitsPerCSSPixel = 1

// PixelsPerDIP answers how many window units the page's CSS pixel takes, given the page's
// devicePixelRatio. On Linux and macOS the window is sized in DIP, so the display's backing scale in
// devicePixelRatio is the toolkit's business and every CSS pixel is one unit.
func PixelsPerDIP(float64) float64 { return unitsPerCSSPixel }
