package main

// The facade the page calls. Every method runs one use case, then does what the window needs
// afterwards. Methods that can be refused answer an error, which rejects the page's promise.

import (
	"context"
	"io"
	"sync/atomic"

	"golang.org/x/sys/windows"

	"github.com/oernster/timestrip/internal/application"
	"github.com/oernster/timestrip/internal/domain/clock"
	"github.com/oernster/timestrip/internal/domain/placement"
	"github.com/oernster/timestrip/internal/domain/settings"
	"github.com/oernster/timestrip/internal/infrastructure/desktop"
)

// Events the page listens for.
const (
	eventRefresh      = "refresh"
	eventOpenSettings = "open-settings"
)

// Which part of Settings an open-settings event asks for.
const (
	openAtSettings = "settings"
	openAtAddClock = "add-clock"
)

// App is the facade Wails binds.
type App struct {
	service  *application.Service
	desktop  *desktop.Desktop
	log      io.Writer
	settings placement.Size

	ctx          context.Context
	strip        windows.HWND
	trayUp       atomic.Bool
	visible      atomic.Bool
	quitting     atomic.Bool
	settingsOpen atomic.Bool
	scrolls      atomic.Bool
}

// newApp answers the facade over service, reporting on desktop, with Settings drawn at settings DIP.
func newApp(service *application.Service, desk *desktop.Desktop, log io.Writer, settingsSize placement.Size) *App {
	return &App{service: service, desktop: desk, log: log, settings: settingsSize}
}

// Snapshot answers what the strip shows now.
func (a *App) Snapshot() snapshotDTO {
	return snapshotOf(a.service.Snapshot(), a.scrolls.Load(), desktop.DragThreshold())
}

// AddClock adds a clock for zone and answers its id (FR-301).
func (a *App) AddClock(zone string) (string, error) {
	id, err := a.service.AddClock(zone)
	a.contentChanged()
	return id, err
}

// RenameClock sets a clock's label (FR-303).
func (a *App) RenameClock(id, label string) error { return a.service.RenameClock(id, label) }

// RezoneClock sets a clock's zone (FR-304).
func (a *App) RezoneClock(id, zone string) error { return a.service.RezoneClock(id, zone) }

// RemoveClock removes a clock; the page has already asked (FR-305).
func (a *App) RemoveClock(id string) error {
	err := a.service.RemoveClock(id)
	a.contentChanged()
	return err
}

// MoveClock moves a clock steps places (FR-306).
func (a *App) MoveClock(id string, steps int) error { return a.service.MoveClock(id, steps) }

// SearchPlaces answers the places matching query (FR-302).
func (a *App) SearchPlaces(query string) []placeDTO { return placesOf(a.service.SearchPlaces(query)) }

// SetStyle chooses digital or analogue (FR-601).
func (a *App) SetStyle(style string) error {
	err := a.service.SetStyle(settings.Style(style))
	a.contentChanged()
	return err
}

// SetFormat chooses 12-hour or 24-hour (FR-206).
func (a *App) SetFormat(format string) error { return a.service.SetFormat(clock.Format(format)) }

// SetOrientation chooses horizontal or vertical (FR-103).
func (a *App) SetOrientation(orientation string) error {
	err := a.service.SetOrientation(settings.Orientation(orientation))
	a.contentChanged()
	return err
}

// SetTheme chooses system, light or dark (FR-606).
func (a *App) SetTheme(theme string) error { return a.service.SetTheme(settings.Theme(theme)) }

// SetAlwaysOnTop turns Always on Top on or off and applies it at once (FR-505).
func (a *App) SetAlwaysOnTop(on bool) error {
	err := a.service.SetAlwaysOnTop(on)
	a.applyAlwaysOnTop()
	return err
}

// StartWithWindows answers whether the Start with Windows value is present (FR-605).
func (a *App) StartWithWindows() (bool, error) { return a.service.StartWithWindows() }

// SetStartWithWindows writes or removes the Start with Windows value (FR-605).
func (a *App) SetStartWithWindows(on bool) error { return a.service.SetStartWithWindows(on) }

// DismissNotices clears the notices the user has read.
func (a *App) DismissNotices() { a.service.DismissNotices() }

// ShowContextMenu shows the strip's right-click menu as a native menu at the cursor (FR-108).
func (a *App) ShowContextMenu() { a.desktop.ShowMenu(a.service.ContextMenu()) }

// OpenSettings turns the window into the Settings surface, centred on the strip's display (CON-6).
func (a *App) OpenSettings() error {
	a.settingsOpen.Store(true)
	at, err := desktop.Position(a.strip)
	if err != nil {
		return err
	}
	arranged, err := a.service.Centred(at, a.settings)
	if err != nil {
		return err
	}
	return desktop.Place(a.strip, arranged.At, arranged.Size)
}

// CloseSettings returns the window to the strip, where it was last left (CON-6, FR-405).
func (a *App) CloseSettings() error {
	a.settingsOpen.Store(false)
	return a.placeLaunched()
}

// Hide hides the strip (FR-504).
func (a *App) Hide() { a.hide() }

// contentChanged fits the strip to what it now holds, keeping its corner (FR-104, FR-105). While
// Settings is open the window is Settings, so the strip is fitted when it closes instead.
func (a *App) contentChanged() {
	if a.settingsOpen.Load() || a.strip == 0 {
		return
	}
	a.rearrange()
}
