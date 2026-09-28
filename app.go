package main

// The facade the page calls. Every method runs one use case, then does what the window needs
// afterwards. Methods that can be refused answer an error, which rejects the page's promise.

import (
	"context"
	"fmt"
	"io"
	"sync/atomic"

	"github.com/oernster/timeribbon/internal/application"
	"github.com/oernster/timeribbon/internal/domain/clock"
	"github.com/oernster/timeribbon/internal/domain/placement"
	"github.com/oernster/timeribbon/internal/domain/settings"
	"github.com/oernster/timeribbon/internal/infrastructure/desktop"
	"github.com/oernster/timeribbon/internal/product"
)

// Events the page listens for.
const (
	eventRefresh   = "refresh"
	eventOpenPanel = "open-panel"
)

// Which panel an open-panel event asks for: Settings, Settings opened on the place search, About,
// Licence or an update check's outcome, which travels with it (CON-6, FR-508, FR-509).
const (
	openAtSettings = "settings"
	openAtAddClock = "add-clock"
	openAtAbout    = "about"
	openAtLicence  = "licence"
	openAtUpdate   = "update"
)

// ribbonService is what the facade asks of the application layer: application.Service in
// production, a scripted stand-in in the facade's tests, which read what the facade decided with
// each answer.
type ribbonService interface {
	Snapshot() application.Snapshot
	Settings() settings.Settings
	AddClock(zone string) (string, error)
	RenameClock(id, label string) error
	RezoneClock(id, zone string) error
	RemoveClock(id string) error
	SearchPlaces(query string) []application.Place
	SetStyle(style settings.Style) error
	SetSize(size settings.Size) error
	SetColour(colour settings.Colour) error
	SetFormat(format clock.Format) error
	SetDateFormat(dateFormat clock.DateFormat) error
	SetOrientation(orientation settings.Orientation) error
	SetTheme(theme settings.Theme) error
	SetAlwaysOnTop(on bool) error
	StartWithWindows() (bool, error)
	SetStartWithWindows(on bool) error
	DismissNotices()
	SetScrollbar(dip int) error
	SetPixelsPerDIP(scale float64) error
	ContextMenu() []application.MenuItem
	CloseRequested() application.MenuAction
	Launch() (application.Arrangement, error)
	Rearrange(at placement.Point) (application.Arrangement, error)
	Moved(at placement.Point) (application.Arrangement, error)
	ToEdge(at placement.Point, edge placement.Edge) (application.Arrangement, error)
	Centred(at placement.Point, size placement.Size) (application.Arrangement, error)
	CheckForUpdate(ctx context.Context, manual bool) application.UpdateStatus
	SkipUpdate(version string) error
}

// App is the facade Wails binds.
type App struct {
	service ribbonService
	desktop *desktop.Desktop
	log     io.Writer
	panel   placement.Size

	// The facade's calls into Wails and the desktop. Each is a field so a test can stand in for it
	// and read what the facade did; newApp points them at the real calls in wails_calls.go and
	// window_life.go.
	emit       func(event string, data ...any)
	showWindow func()
	hideWindow func()
	quit       func()
	setOnTop   func(on bool)
	browse     func(address string) error
	showMenu   func(items []application.MenuItem)
	position   func() (placement.Point, error)
	place      func(at placement.Point, size placement.Size) error

	ctx       context.Context
	ribbon    desktop.Window
	trayUp    atomic.Bool
	visible   atomic.Bool
	quitting  atomic.Bool
	panelOpen atomic.Bool
	scrolls   atomic.Bool

	// updates holds the update check's timing and the outcome it last offered (FR-509).
	updates updateWatch
}

// newApp answers the facade over service, reporting on desktop, with every panel drawn at panel DIP.
func newApp(service ribbonService, desk *desktop.Desktop, log io.Writer, panelSize placement.Size) *App {
	built := &App{
		service: service, desktop: desk, log: log, panel: panelSize,
		updates: updateWatch{delay: updateCheckDelay, every: updateCheckEvery},
	}
	built.emit = built.emitToWails
	built.showWindow = built.showInWails
	built.hideWindow = built.hideInWails
	built.quit = built.quitWails
	built.setOnTop = built.setOnTopInWails
	built.browse = desktop.OpenInBrowser
	built.showMenu = desk.ShowMenu
	built.position = built.ribbonPosition
	built.place = built.placeRibbon
	return built
}

// Snapshot answers what the ribbon shows now.
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
func (a *App) RenameClock(id, label string) error {
	return a.refitted(a.service.RenameClock(id, label))
}

// RezoneClock sets a clock's zone (FR-304).
func (a *App) RezoneClock(id, zone string) error { return a.refitted(a.service.RezoneClock(id, zone)) }

// RemoveClock removes a clock; the page has already asked (FR-305).
func (a *App) RemoveClock(id string) error {
	err := a.service.RemoveClock(id)
	a.contentChanged()
	return err
}

// SearchPlaces answers the places matching query (FR-302).
func (a *App) SearchPlaces(query string) []placeDTO { return placesOf(a.service.SearchPlaces(query)) }

// SetStyle chooses digital or analogue (FR-601).
func (a *App) SetStyle(style string) error {
	err := a.service.SetStyle(settings.Style(style))
	a.contentChanged()
	return err
}

// SetSize chooses large or small cells (FR-610).
func (a *App) SetSize(size string) error {
	err := a.service.SetSize(settings.Size(size))
	a.contentChanged()
	return err
}

// SetColour chooses the colour scheme (FR-611).
func (a *App) SetColour(colour string) error {
	return a.refitted(a.service.SetColour(settings.Colour(colour)))
}

// SetFormat chooses 12-hour or 24-hour (FR-206).
func (a *App) SetFormat(format string) error {
	return a.refitted(a.service.SetFormat(clock.Format(format)))
}

// SetDateFormat chooses how every date is written (FR-612).
func (a *App) SetDateFormat(dateFormat string) error {
	return a.refitted(a.service.SetDateFormat(clock.DateFormat(dateFormat)))
}

// SetOrientation chooses horizontal or vertical (FR-103), then puts the ribbon against that
// orientation's home edge (FR-409). A choice that did not take, as one the setting does not offer,
// fits the ribbon where it stands. One whose save failed has still taken, so it moves.
func (a *App) SetOrientation(orientation string) error {
	chosen := settings.Orientation(orientation)
	err := a.service.SetOrientation(chosen)
	edge, known := settings.HomeEdge(chosen)
	if !known || a.service.Settings().Orientation != chosen {
		a.contentChanged()
		return err
	}
	a.toEdge(edge)
	return err
}

// SetTheme chooses system, light or dark (FR-606).
func (a *App) SetTheme(theme string) error {
	return a.refitted(a.service.SetTheme(settings.Theme(theme)))
}

// SetAlwaysOnTop turns Always on Top on or off and applies it at once (FR-505).
func (a *App) SetAlwaysOnTop(on bool) error {
	err := a.service.SetAlwaysOnTop(on)
	a.applyAlwaysOnTop()
	return a.refitted(err)
}

// StartWithWindows answers whether the Start with Windows value is present (FR-605).
func (a *App) StartWithWindows() (bool, error) { return a.service.StartWithWindows() }

// SetStartWithWindows writes or removes the Start with Windows value (FR-605).
func (a *App) SetStartWithWindows(on bool) error { return a.service.SetStartWithWindows(on) }

// DismissNotices clears the notices the user has read, then fits the ribbon without their cells.
func (a *App) DismissNotices() {
	a.service.DismissNotices()
	a.contentChanged()
}

// SetScrollbar takes the thickness in DIP of the scroll bar the page draws, which it measures once
// it has loaded, then fits the ribbon with room for it (FR-106).
func (a *App) SetScrollbar(dip int) error { return a.refitted(a.service.SetScrollbar(dip)) }

// SetPixelRatio takes the page's devicePixelRatio, which it reports once it has loaded and again
// whenever it changes, then fits the window to the page as it is really drawn. Windows' text size
// enlarges the page without changing the display's DPI, so the DPI alone left the page cut off.
func (a *App) SetPixelRatio(ratio float64) error {
	return a.refitted(a.service.SetPixelsPerDIP(desktop.PixelsPerDIP(ratio)))
}

// ShowContextMenu shows the ribbon's right-click menu as a native menu at the cursor (FR-108).
func (a *App) ShowContextMenu() { a.showMenu(a.service.ContextMenu()) }

// OpenPanel turns the window into a panel (Settings, About or Licence), centred on the ribbon's
// display (CON-6).
func (a *App) OpenPanel() error {
	a.panelOpen.Store(true)
	at, err := a.position()
	if err != nil {
		return err
	}
	arranged, err := a.service.Centred(at, a.panel)
	if err != nil {
		return err
	}
	return a.place(arranged.At, arranged.Size)
}

// ClosePanel returns the window to the ribbon, where it was last left (CON-6, FR-405).
func (a *App) ClosePanel() error {
	a.panelOpen.Store(false)
	return a.placeLaunched()
}

// OpenDonation hands the donation page to the desktop's browser. The application never fetches it,
// so the button adds no network request to the update check's one (NFR-S-1). Where Windows
// cannot open it, the refusal says why and gives the address, so it can still be reached by hand.
func (a *App) OpenDonation() error {
	if err := a.browse(product.DonateURL); err != nil {
		return fmt.Errorf("your browser could not be opened on the donation page (%w). Windows may have no default browser set; the page is %s", err, product.DonateURL)
	}
	return nil
}

// Hide hides the ribbon (FR-504).
func (a *App) Hide() { a.hide() }

// refitted fits the ribbon after a change, then answers err. A change whose save failed raises a
// notice, which is one more cell to fit (FR-707); one that saved may have ended an earlier notice.
func (a *App) refitted(err error) error {
	a.contentChanged()
	return err
}

// contentChanged fits the ribbon to what it now holds, keeping its corner (FR-104, FR-105). While a
// panel is open the window is that panel, so the ribbon is fitted when it closes instead.
func (a *App) contentChanged() {
	if a.panelOpen.Load() || a.ribbon == 0 {
		return
	}
	a.rearrange()
}
