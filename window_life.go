package main

// The window's own life: startup, showing, hiding, closing and what the desktop reports.

import (
	"context"
	"fmt"

	"github.com/oernster/timeribbon/internal/application"
	"github.com/oernster/timeribbon/internal/domain/placement"
	"github.com/oernster/timeribbon/internal/infrastructure/desktop"
	"github.com/oernster/timeribbon/internal/product"
)

// startup takes the strip off the taskbar and puts it in place while it is still hidden, then
// starts listening to the desktop. Nothing here ends the run: a failure is logged and the strip
// opens wherever Wails put it.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	strip, err := desktop.FindStrip(product.StripClass)
	if err != nil {
		a.report("finding the strip", err)
		return
	}
	a.strip = strip
	a.report("hiding the taskbar button", desktop.HideFromTaskbar(strip))
	a.report("keeping the strip on its displays", desktop.KeepOnDisplays(strip, a.log))
	a.desktop.Watch(strip)
	a.report("placing the strip", a.placeLaunched())
	a.applyAlwaysOnTop()
	go a.listen()
}

// domReady shows the strip once the page has drawn, so it never appears blank.
func (a *App) domReady(context.Context) { a.show() }

// beforeClose answers a request to close the strip, such as Alt+F4: it hides the strip and the
// application keeps running (FR-507). An Exit already decided passes through, as does any close
// while there is no tray icon to bring the strip back from.
func (a *App) beforeClose(context.Context) bool {
	if a.quitting.Load() || !a.trayUp.Load() {
		return false
	}
	if a.service.CloseRequested() == application.ActionHide {
		a.hide()
	}
	return true
}

func (a *App) shutdown(context.Context) { a.desktop.Stop() }

// secondInstance answers a second launch by showing the strip that is already running (FR-506).
func (a *App) secondInstance() { a.show() }

// listen acts on what the desktop reports until it stops. A panic in one event is logged and the
// next is still heard, so one fault cannot leave a strip that reacts to nothing.
func (a *App) listen() {
	for event := range a.desktop.Events() {
		a.handleSafely(event)
	}
}

func (a *App) handleSafely(event desktop.Event) {
	defer func() {
		if failure := recover(); failure != nil {
			fmt.Fprintf(a.log, "recovered from %v while handling desktop event %d\n", failure, event.Kind)
		}
	}()
	switch event.Kind {
	case desktop.EventMenu:
		a.act(event.Action)
	case desktop.EventIconClicked:
		a.toggle()
	case desktop.EventMoveEnded:
		a.moved()
	case desktop.EventDisplayChanged:
		fmt.Fprintln(a.log, "the displays changed")
		a.rearrange()
		a.emit(eventRefresh)
	case desktop.EventTimeChanged, desktop.EventResumed:
		fmt.Fprintf(a.log, "desktop event %d: refreshing\n", event.Kind)
		a.emit(eventRefresh)
	}
}

// act carries out a menu action from the tray or the strip's own menu.
func (a *App) act(action application.MenuAction) {
	switch action {
	case application.ActionShow:
		a.show()
	case application.ActionHide:
		a.hide()
	case application.ActionAddClock:
		a.show()
		a.emit(eventOpenPanel, openAtAddClock)
	case application.ActionSettings:
		a.show()
		a.emit(eventOpenPanel, openAtSettings)
	case application.ActionAbout:
		a.show()
		a.emit(eventOpenPanel, openAtAbout)
	case application.ActionLicence:
		a.show()
		a.emit(eventOpenPanel, openAtLicence)
	case application.ActionAlwaysOnTop:
		a.report("changing Always on top", a.SetAlwaysOnTop(!a.service.Settings().AlwaysOnTop))
		a.emit(eventRefresh)
	case application.ActionExit:
		a.quitting.Store(true)
		if a.ctx != nil {
			a.quit()
		}
	default:
		a.actOnChoice(action)
	}
}

// actOnChoice carries out a Position, Style or Orientation item (FR-108, FR-408, FR-409), then has
// the page redraw, since a choice made from a menu is one the page did not make.
func (a *App) actOnChoice(action application.MenuAction) {
	if edge, ok := application.EdgeOf(action); ok {
		a.toEdge(edge)
		return
	}
	if style, ok := application.StyleOf(action); ok {
		a.report("changing the style", a.SetStyle(string(style)))
		a.emit(eventRefresh)
		return
	}
	if orientation, ok := application.OrientationOf(action); ok {
		a.report("changing the orientation", a.SetOrientation(string(orientation)))
		a.emit(eventRefresh)
	}
}

// toEdge puts the strip against edge of its display and shows it there (FR-408). While a panel is
// open the window is that panel, so the place is kept and the strip goes there as the panel closes.
// Before startup has found the strip there is nothing to move.
func (a *App) toEdge(edge placement.Edge) {
	if a.strip == 0 {
		return
	}
	at, err := a.position()
	if err != nil {
		a.report("reading where the strip is", err)
		return
	}
	arranged, err := a.service.ToEdge(at, edge)
	if err != nil {
		a.report("putting the strip against an edge", err)
		return
	}
	if a.panelOpen.Load() {
		return
	}
	a.scrolls.Store(arranged.Scrolls)
	a.report("placing the strip", a.place(arranged.At, arranged.Size))
	a.show()
}

// moved records where a drag left the strip, putting it back onto a display if the drag left part
// of it off every one (FR-404, FR-406). A move of a panel is not the strip's.
func (a *App) moved() {
	if a.panelOpen.Load() {
		return
	}
	at, err := a.position()
	if err != nil {
		a.report("reading where the strip was left", err)
		return
	}
	arranged, err := a.service.Moved(at)
	a.report("recording where the strip was left", err)
	if err != nil {
		// The placement could not be saved, which raised a notice: fit the strip where it stands,
		// its new cell included, rather than leave it wherever the drag let go.
		a.rearrange()
		return
	}
	a.scrolls.Store(arranged.Scrolls)
	a.report("placing the strip", a.place(arranged.At, arranged.Size))
}

// rearrange fits the strip where it stands (FR-104, FR-406).
func (a *App) rearrange() {
	if a.panelOpen.Load() {
		return
	}
	at, err := a.position()
	if err != nil {
		a.report("reading where the strip is", err)
		return
	}
	arranged, err := a.service.Rearrange(at)
	if err != nil {
		a.report("fitting the strip", err)
		return
	}
	a.scrolls.Store(arranged.Scrolls)
	a.report("placing the strip", a.place(arranged.At, arranged.Size))
}

// placeLaunched puts the strip where it was last left (FR-405).
func (a *App) placeLaunched() error {
	arranged, err := a.service.Launch()
	if err != nil {
		return err
	}
	a.scrolls.Store(arranged.Scrolls)
	return a.place(arranged.At, arranged.Size)
}

// stripPosition and placeStrip are the production position and place: the strip's window as the
// desktop reports and moves it.
func (a *App) stripPosition() (placement.Point, error) { return desktop.Position(a.strip) }

func (a *App) placeStrip(at placement.Point, size placement.Size) error {
	return desktop.Place(a.strip, at, size)
}

func (a *App) applyAlwaysOnTop() {
	if a.ctx != nil {
		a.setOnTop(a.service.Settings().AlwaysOnTop)
	}
}

func (a *App) show() {
	if a.ctx == nil {
		return
	}
	a.showWindow()
	a.visible.Store(true)
	a.emit(eventRefresh)
}

func (a *App) hide() {
	if a.ctx == nil {
		return
	}
	a.hideWindow()
	a.visible.Store(false)
}

func (a *App) toggle() {
	if a.visible.Load() {
		a.hide()
		return
	}
	a.show()
}

// report writes a failure to the log; nothing when there was none.
func (a *App) report(doing string, err error) {
	if err != nil {
		fmt.Fprintf(a.log, "%s: %v\n", doing, err)
	}
}
