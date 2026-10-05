package main

// The sun map's half of the facade (FR-901 to FR-913). The map shares the ribbon's window: while it
// shows, the window is the ribbon and the map together; the page lays the two out inside it.
// Every placement is decided for the ribbon alone; this file turns it into the window's and back.

import (
	"github.com/oernster/timeribbon/ribbonkit/application/arranger"
	"github.com/oernster/timeribbon/ribbonkit/domain/placement"
)

// windowOf answers the window's place and size for the ribbon arranged as full, shown in full: the
// ribbon with its map where one shows. offset is where the ribbon's corner lies inside the window.
func windowOf(full arranger.Arrangement) (at placement.Point, size placement.Size, offset placement.Point) {
	ribbon := placement.Rect{Left: full.At.X, Top: full.At.Y, Right: full.At.X + full.Size.Width, Bottom: full.At.Y + full.Size.Height}
	whole := ribbon
	if full.Map != (placement.Rect{}) {
		whole = placement.Rect{
			Left: min(ribbon.Left, full.Map.Left), Top: min(ribbon.Top, full.Map.Top),
			Right: max(ribbon.Right, full.Map.Right), Bottom: max(ribbon.Bottom, full.Map.Bottom),
		}
	}
	at = placement.Point{X: whole.Left, Y: whole.Top}
	return at, placement.Size{Width: whole.Width(), Height: whole.Height()}, placement.Point{X: ribbon.Left - whole.Left, Y: ribbon.Top - whole.Top}
}

// ribbonFromWindow answers where the ribbon stands for a window at at showing the full ribbon,
// which a drag moves as one.
func (a *App) ribbonFromWindow(at placement.Point) placement.Point {
	a.unpin.guard.Lock()
	_, _, offset := windowOf(a.unpin.full)
	a.unpin.guard.Unlock()
	return placement.Point{X: at.X + offset.X, Y: at.Y + offset.Y}
}

// SetSunMap turns the sun map on or off (FR-901), then fits the window to the ribbon with or without
// it and has the page draw what the window now holds.
func (a *App) SetSunMap(on bool) error { return a.redrawn(a.refitted(a.service.SetSunMap(on))) }

// TogglePullOut opens a vertical ribbon's pull out when closed and closes it when open (FR-903), as
// its handle is clicked. The window changes size, so the page is told to draw it again: without that
// it kept the closed layout in a window grown for the map (Oliver, 2026-09-29).
func (a *App) TogglePullOut() error {
	return a.redrawn(a.refitted(a.service.SetPullOut(!a.service.Settings().PullOut)))
}

// placeShaped cuts the window to parts (FR-913), then puts it at at, size across. The cut comes
// first, in the new window's pixels, so a window growing for the map never shows the bands it is cut
// from. A failed cut is written to the log and the window is placed all the same, as a rectangle.
func (a *App) placeShaped(at placement.Point, size placement.Size, parts []placement.Rect) error {
	a.report("cutting the window to the ribbon and its map", a.shape(parts))
	a.lastPlaced.note(at)
	return a.place(at, size)
}

// placeWhole puts the window at at, size across, keeping all of it: the tab and a panel.
func (a *App) placeWhole(at placement.Point, size placement.Size) error {
	return a.placeShaped(at, size, placement.Shape(size, placement.Rect{}, placement.Rect{}, false))
}

// redrawn tells the page to take a fresh snapshot, then answers err.
func (a *App) redrawn(err error) error {
	a.emit(eventRefresh)
	return err
}

// mapLayout answers where the page draws the ribbon and its map inside the window, in the window's
// pixels: the map shown only while the full ribbon is, never with the tab or a panel (FR-910). An
// opening ribbon counts as shown while the page draws it, as collapsed agrees, since the window grows
// round what was drawn then (FR-615).
func (a *App) mapLayout() (side placement.Edge, ribbon, sunMap placement.Rect, shown bool) {
	a.unpin.guard.Lock()
	full := a.unpin.full
	open := (a.unpin.shownOpen || a.unpin.drawing) && !a.panelOpen.Load()
	a.unpin.guard.Unlock()
	at, _, offset := windowOf(full)
	ribbon = placement.Rect{Left: offset.X, Top: offset.Y, Right: offset.X + full.Size.Width, Bottom: offset.Y + full.Size.Height}
	if !open || full.Map == (placement.Rect{}) {
		return full.MapSide, ribbon, placement.Rect{}, false
	}
	sunMap = placement.Rect{Left: full.Map.Left - at.X, Top: full.Map.Top - at.Y, Right: full.Map.Right - at.X, Bottom: full.Map.Bottom - at.Y}
	return full.MapSide, ribbon, sunMap, true
}
