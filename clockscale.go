package main

// PreviewScale draws the ribbon at percent while its grip is dragged, keeping nothing, then fits
// the window to it and tells the page to draw again (FR-623).
func (a *App) PreviewScale(percent int) error {
	return a.redrawn(a.refitted(a.service.PreviewScale(percent)))
}

// SetScale keeps the scale the grip's drag ended at, then fits the window to it and tells the page
// to draw again (FR-623).
func (a *App) SetScale(percent int) error {
	return a.redrawn(a.refitted(a.service.SetScale(percent)))
}
