package desktop

import (
	"errors"
	"fmt"
	"io"
	"runtime/cgo"
	"sync"
	"time"

	"github.com/oernster/timeribbon/internal/application"
	"github.com/oernster/timeribbon/internal/domain/placement"
)

// errNoTray is answered by Start until the Linux tray icon is built: without it, closing the ribbon
// exits, as the facade already does wherever the icon could not start.
var errNoTray = errors.New("the tray icon is not yet built on Linux")

// moveSettle is how long the ribbon must stand still before a move counts as ended. The window
// manager carries a drag through without saying when the button is let go, so the end is the
// moment the position stops changing.
const moveSettle = 300 * time.Millisecond

// Desktop hears the desktop on Linux: the end of the ribbon's moves, changes of display and jumps
// of the clock, plus the ribbon's own menu. GTK is not open when Start runs, so everything that
// needs GTK waits for Watch, which runs once the ribbon's window exists.
type Desktop struct {
	menu   func() []application.MenuItem
	events chan Event
	log    io.Writer
	stop   chan struct{}

	guard  sync.Mutex
	closed bool
	ribbon Window
	lastAt placement.Point
	settle *time.Timer

	started sync.Once
	watched sync.Once
	stopped sync.Once
}

// New answers a desktop whose tray menu is menu, reporting failures to log.
func New(menu func() []application.MenuItem, log io.Writer) *Desktop {
	return &Desktop{menu: menu, events: make(chan Event, eventBuffer), log: log, stop: make(chan struct{})}
}

// Events yields what happened. The channel is closed when the desktop stops.
func (d *Desktop) Events() <-chan Event { return d.events }

// Start begins watching the clock, then answers why there is no tray icon.
func (d *Desktop) Start() error {
	d.started.Do(func() {
		go watchClock(d.stop, func() { d.send(Event{Kind: EventTimeChanged}) })
	})
	return errNoTray
}

// Stop ends the watching and closes the events.
func (d *Desktop) Stop() {
	d.stopped.Do(func() {
		close(d.stop)
		d.guard.Lock()
		defer d.guard.Unlock()
		d.closed = true
		if d.settle != nil {
			d.settle.Stop()
		}
		close(d.events)
	})
}

// Watch names the ribbon's window, so the end of its moves is reported, then hears the changes of
// display. It runs once GTK is open.
func (d *Desktop) Watch(ribbon Window) {
	d.watched.Do(func() {
		d.guard.Lock()
		d.ribbon = ribbon
		d.guard.Unlock()
		if err := watchWindow(ribbon, cgo.NewHandle(d)); err != nil {
			fmt.Fprintf(d.log, "desktop: watching the ribbon: %v\n", err)
		}
	})
}

// ShowMenu shows items as a native menu at the pointer, from any goroutine: the ribbon's right-click
// menu (FR-108). The choice arrives as an EventMenu. Before the ribbon is found there is nothing to
// show it over.
func (d *Desktop) ShowMenu(items []application.MenuItem) {
	d.guard.Lock()
	ribbon := d.ribbon
	d.guard.Unlock()
	if ribbon == 0 {
		return
	}
	if err := popUp(ribbon, d, items); err != nil {
		fmt.Fprintf(d.log, "desktop: showing the menu: %v\n", err)
	}
}

// moved hears the ribbon standing at at. A position the ribbon was placed at is not a move; it also
// cancels any wait already begun, since on the way there the window manager may stand it somewhere
// else for a moment, as when it grows before it moves. Any other position starts the wait for it
// to settle, the wait starting over while it keeps moving.
func (d *Desktop) moved(at placement.Point) {
	d.guard.Lock()
	defer d.guard.Unlock()
	placedTo, placedOK := placedAt(d.ribbon)
	if d.closed || at == d.lastAt {
		return
	}
	d.lastAt = at
	if d.settle != nil {
		d.settle.Stop()
	}
	if placedOK && placedTo == at {
		return
	}
	d.settle = time.AfterFunc(moveSettle, func() { d.send(Event{Kind: EventMoveEnded}) })
}

// send hands event on without waiting: one nobody is reading is dropped and said so, since the
// desktop calls in on GTK's own thread, which must never block.
func (d *Desktop) send(event Event) {
	d.guard.Lock()
	defer d.guard.Unlock()
	if d.closed {
		return
	}
	select {
	case d.events <- event:
	default:
		fmt.Fprintf(d.log, "desktop: event %d dropped, nothing is reading\n", event.Kind)
	}
}
