package desktop

// PixelsPerDIP answers how many window pixels the page's CSS pixel takes, given the page's
// devicePixelRatio. On Windows the window is sized in physical pixels and WebView2 draws the page at
// its rasterization scale, the display's scale combined with the user's text size, which is what
// devicePixelRatio reports; so the two are the same.
func PixelsPerDIP(pageRatio float64) float64 { return pageRatio }
