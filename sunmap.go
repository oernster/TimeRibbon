package main

// The sun map's half of the facade (FR-901 to FR-910). The map shares the ribbon's window: while it
// shows, the window is the ribbon and the map together; the page lays the two out inside it.
// Every placement is decided for the ribbon alone; this file turns it into the window's and back.

import (
	"github.com/oernster/timeribbon/internal/application"
	"github.com/oernster/timeribbon/internal/domain/placement"
)

// windowOf answers the window's place and size for the ribbon arranged as full, shown in full: the
// ribbon with its map where one shows. offset is where the ribbon's corner lies inside the window.
func windowOf(full application.Arrangement) (at placement.Point, size placement.Size, offset placement.Point) {
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
// it.
func (a *App) SetSunMap(on bool) error { return a.refitted(a.service.SetSunMap(on)) }

// TogglePullOut opens a vertical ribbon's pull out when closed and closes it when open (FR-903), as
// its handle is clicked.
func (a *App) TogglePullOut() error {
	return a.refitted(a.service.SetPullOut(!a.service.Settings().PullOut))
}

// mapLayout answers where the page draws the ribbon and its map inside the window, in the window's
// pixels: the map shown only while the full ribbon is, never with the tab or a panel (FR-910).
func (a *App) mapLayout() (side placement.Edge, ribbon, sunMap placement.Rect, shown bool) {
	a.unpin.guard.Lock()
	full := a.unpin.full
	open := a.unpin.shownOpen && !a.panelOpen.Load()
	a.unpin.guard.Unlock()
	at, _, offset := windowOf(full)
	ribbon = placement.Rect{Left: offset.X, Top: offset.Y, Right: offset.X + full.Size.Width, Bottom: offset.Y + full.Size.Height}
	if !open || full.Map == (placement.Rect{}) {
		return full.MapSide, ribbon, placement.Rect{}, false
	}
	sunMap = placement.Rect{Left: full.Map.Left - at.X, Top: full.Map.Top - at.Y, Right: full.Map.Right - at.X, Bottom: full.Map.Bottom - at.Y}
	return full.MapSide, ribbon, sunMap, true
}
