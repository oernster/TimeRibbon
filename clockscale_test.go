package main

import "testing"

// FR-623: the grip's page cannot take a snapshot mid-drag by itself, so a preview and the kept
// scale each tell it to draw again, as well as fitting the window (the fit is proved in
// TestEveryChangeFitsTheRibbonAndAnswersTheServicesError).
func TestAChangeOfScaleTellsThePageToDrawAgain(t *testing.T) {
	for name, change := range map[string]func(*App) error{
		"PreviewScale": func(app *App) error { return app.PreviewScale(150) },
		"SetScale":     func(app *App) error { return app.SetScale(150) },
	} {
		app, _, seen, _ := newTestApp(t)
		if err := change(app); err != nil || !seen.sawEvent(eventRefresh) {
			t.Errorf("%s answered %v; refresh sent %v", name, err, seen.sawEvent(eventRefresh))
		}
	}
}
