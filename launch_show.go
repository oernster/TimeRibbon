package main

import (
	"sync/atomic"
	"time"
)

// sizeWait is how long a launched ribbon whose page is ready waits for the page's reports that size
// it before it is shown anyway, so a page that never reports still shows a ribbon.
const sizeWait = time.Second

// launchShow holds what the first showing of a launched ribbon waits for: the page ready to show,
// its scale applied to the window and its widest text measured, the two reports that size the
// window after it loads. Shown at domReady alone, the window appeared at one size, then grew; at a
// fractional KDE scale WebKitGTK then often kept painting the size it was first shown at, leaving
// the ribbon cut off until the page next changed (measured 2026-10-02 on Plasma at 150 percent:
// shown at once, 8 of 16 launches were cut off to the first size; shown once scaled but before the
// widths, 13 of 16 were cut to the width of that moment).
type launchShow struct {
	ready    atomic.Bool
	scaled   atomic.Bool
	measured atomic.Bool
	shown    atomic.Bool
}

// pageReady records that the page can be shown, then shows the ribbon if it is sized; else it waits
// sizeWait for that.
func (a *App) pageReady() {
	a.launch.ready.Store(true)
	if !a.showLaunched() {
		a.after(sizeWait, a.sizeDue)
	}
}

// pageScaled records that the page's scale has been applied to the window.
func (a *App) pageScaled() {
	a.launch.scaled.Store(true)
	a.showLaunched()
}

// pageMeasured records that the page's widest text has been applied to the window.
func (a *App) pageMeasured() {
	a.launch.measured.Store(true)
	a.showLaunched()
}

// sizeDue shows a ready ribbon whose page has not finished sizing it within sizeWait.
func (a *App) sizeDue() {
	a.launch.scaled.Store(true)
	a.launch.measured.Store(true)
	a.showLaunched()
}

// showLaunched shows the launched ribbon once, when the page is ready and has sized it; it answers
// whether it has been shown.
func (a *App) showLaunched() bool {
	if !a.launch.ready.Load() || !a.launch.scaled.Load() || !a.launch.measured.Load() {
		return false
	}
	if a.launch.shown.CompareAndSwap(false, true) {
		a.show()
	}
	return true
}
