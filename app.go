package main

// The facade the page calls. The ribbon's window is ribbonkit's, embedded here so its methods are page
// API alongside these; what is TimeRibbon's own (its clocks, their style, size and formats, the sun
// map and the measuring of their text) is here. Every method runs one use case, then has the window
// do what it needs afterwards. Methods that can be refused answer an error, which rejects the page's
// promise.

import (
	"github.com/oernster/ribbonkit/application/menus"
	"github.com/oernster/ribbonkit/domain/localtime"
	"github.com/oernster/ribbonkit/ui/window"
	"github.com/oernster/timeribbon/internal/application"
	"github.com/oernster/timeribbon/internal/domain/clock"
	"github.com/oernster/timeribbon/internal/domain/settings"
	"github.com/oernster/timeribbon/internal/product"
)

// openAtAddClock asks the page to open Settings on the place search (FR-301).
const openAtAddClock = "add-clock"

// ribbonService is what TimeRibbon's own half of the facade asks of the application layer:
// application.Service in production, a scripted stand-in in the facade's tests. The window asks for
// the rest through window.Service.
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
	SetFormat(format localtime.Format) error
	SetDateFormat(dateFormat clock.DateFormat) error
	SetSunMap(on bool) error
	DismissNotices()
	TextSamples() (times, dates []string)
	SetMeasured(measured application.Measured) error
	SettingsChoices() []menus.Item
}

// windowControl is what TimeRibbon's own half asks of the window: the Control's calls, plus the two
// ribbon choices its menus carry. kitWindow is the real one; the facade's tests stand in for it.
type windowControl interface {
	Refitted(err error) error
	ContentChanged()
	Redraw()
	Redrawn(err error) error
	Report(doing string, err error)
	ShowPanel(panel string)
	PageMeasured()
	Shown() window.Shown
	SetColour(colour string) error
	SetOrientation(orientation string) error
}

// kitWindow is the window's two halves together, as TimeRibbon's own half reaches them.
type kitWindow struct {
	*window.Window
	*window.Control
}

// App is the facade Wails binds.
type App struct {
	*window.Window
	service ribbonService
	control windowControl
}

// newApp answers the facade over service, its window built from config with TimeRibbon's product and
// menu actions, plus the Control the composition root runs it with.
func newApp(service ribbonService, config window.Config) (*App, *window.Control) {
	app := &App{service: service}
	config.Act = app.actOn
	config.Product = productOf()
	shown, control := window.New(config)
	app.Window, app.control = shown, kitWindow{shown, control}
	return app, control
}

// Snapshot answers what the ribbon shows now, the tab and the map's place included (FR-614, FR-910).
func (a *App) Snapshot() snapshotDTO {
	seen := a.control.Shown()
	shown := snapshotOf(a.service.Snapshot(), seen.Scrolls, seen.DragThreshold)
	shown.Collapsed = seen.Collapsed
	shown.SunMap.Side, shown.SunMap.Shown = string(seen.PullOutSide), seen.PullOutShown
	shown.SunMap.Ribbon, shown.SunMap.Map = seen.Ribbon, seen.PullOut
	shown.Choices = choicesOf(a.service.SettingsChoices())
	return shown
}

// AddClock adds a clock for zone and answers its id (FR-301).
func (a *App) AddClock(zone string) (string, error) {
	id, err := a.service.AddClock(zone)
	a.control.ContentChanged()
	return id, err
}

// RenameClock sets a clock's label (FR-303).
func (a *App) RenameClock(id, label string) error {
	return a.control.Refitted(a.service.RenameClock(id, label))
}

// RezoneClock sets a clock's zone (FR-304).
func (a *App) RezoneClock(id, zone string) error {
	return a.control.Refitted(a.service.RezoneClock(id, zone))
}

// RemoveClock removes a clock; the page has already asked (FR-305).
func (a *App) RemoveClock(id string) error {
	err := a.service.RemoveClock(id)
	a.control.ContentChanged()
	return err
}

// SearchPlaces answers the places matching query (FR-302).
func (a *App) SearchPlaces(query string) []placeDTO { return placesOf(a.service.SearchPlaces(query)) }

// SetStyle chooses digital or analogue (FR-601).
func (a *App) SetStyle(style string) error {
	err := a.service.SetStyle(settings.Style(style))
	a.control.ContentChanged()
	return err
}

// SetSize chooses large or small cells (FR-610).
func (a *App) SetSize(size string) error {
	err := a.service.SetSize(settings.Size(size))
	a.control.ContentChanged()
	return err
}

// SetFormat chooses 12-hour or 24-hour (FR-206).
func (a *App) SetFormat(format string) error {
	return a.control.Refitted(a.service.SetFormat(localtime.Format(format)))
}

// SetDateFormat chooses how every date is written (FR-612).
func (a *App) SetDateFormat(dateFormat string) error {
	return a.control.Refitted(a.service.SetDateFormat(clock.DateFormat(dateFormat)))
}

// SetSunMap turns the sun map on or off (FR-901), then fits the window to the ribbon with or without
// it and has the page draw what the window now holds.
func (a *App) SetSunMap(on bool) error {
	return a.control.Redrawn(a.control.Refitted(a.service.SetSunMap(on)))
}

// DismissNotices clears the notices the user has read, then fits the ribbon without their cells.
func (a *App) DismissNotices() {
	a.service.DismissNotices()
	a.control.ContentChanged()
}

// actOn carries out a menu action of TimeRibbon's own (FR-108, FR-301, FR-901), then has the page
// redraw, since a choice made from a menu is one the page did not make. The window hands over every
// action it does not know; one that is not TimeRibbon's either changes nothing.
func (a *App) actOn(action menus.Action) {
	if action == application.ActionAddClock {
		a.control.ShowPanel(openAtAddClock)
		return
	}
	if choose, doing, ok := a.choiceOf(action); ok {
		a.control.Report(doing, choose())
		a.control.Redraw()
	}
}

// choiceOf answers the choice action names with what making it is doing; false where action names
// none of TimeRibbon's choices.
func (a *App) choiceOf(action menus.Action) (choose func() error, doing string, ok bool) {
	if action == application.ActionSunMap {
		return func() error { return a.SetSunMap(!a.service.Settings().SunMap) }, "turning the sun map on or off", true
	}
	if style, ok := application.StyleOf(action); ok {
		return func() error { return a.SetStyle(string(style)) }, "changing the style", true
	}
	if colour, ok := application.ColourOf(action); ok {
		return func() error { return a.control.SetColour(string(colour)) }, "changing the colour", true
	}
	if orientation, ok := application.OrientationOf(action); ok {
		return func() error { return a.control.SetOrientation(string(orientation)) }, "changing the orientation", true
	}
	return nil, "", false
}

// productOf answers what the window says about TimeRibbon, every word from internal/product and the
// terms the LICENSE file itself (FR-607, FR-608).
func productOf() window.Product {
	credits := make([]window.Credit, 0, len(product.Credits))
	for _, credit := range product.Credits {
		credits = append(credits, window.Credit{Name: credit.Name, Licence: credit.Licence, Role: credit.Role})
	}
	return window.Product{
		App: product.App(), WindowClass: product.RibbonClass, Version: product.Version,
		Author: product.Author, Copyright: product.Copyright, Credits: credits,
		Licence: licenceText, DonateURL: product.DonateURL,
	}
}
