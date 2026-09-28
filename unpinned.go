package main

// The unpinned ribbon (FR-613 to FR-618): the facade's half of collapsing to the tab and opening
// from it. When is hover's to decide; this file carries the decision out on the window.

import (
	"fmt"
	"sync"
	"time"

	"github.com/oernster/timeribbon/internal/application"
	"github.com/oernster/timeribbon/internal/domain/hover"
	"github.com/oernster/timeribbon/internal/domain/placement"
)

// unpinned is the hover state with what the window shows because of it.
type unpinned struct {
	guard sync.Mutex
	state hover.State
	// shownOpen is whether the window shows the full ribbon rather than its tab.
	shownOpen bool
	// full is the full ribbon's last arrangement, which the tab is cut from and opening returns to.
	full application.Arrangement
	// holds counts what keeps the ribbon as it is: an open panel, the ribbon's own menu.
	holds    int
	menuHeld bool
	// stop cancels the pending timer; nil when none is pending.
	stop func() bool
}

// pinned answers whether the ribbon is pinned, which leaves it shown in full (FR-613).
func (a *App) pinned() bool { return a.service.Settings().Pinned }

// collapsed answers whether the window is the ribbon's tab, which the page draws as a band (FR-614).
func (a *App) collapsed() bool {
	a.unpin.guard.Lock()
	defer a.unpin.guard.Unlock()
	return !a.unpin.shownOpen && !a.panelOpen.Load()
}

// arrangeWindow places the window for the ribbon arranged as full: the full ribbon, else its tab
// while unpinned and collapsed. Every placement of the ribbon comes through here.
func (a *App) arrangeWindow(full application.Arrangement) error {
	open := a.pinned()
	a.unpin.guard.Lock()
	a.unpin.full = full
	open = open || a.unpin.state.Open()
	a.unpin.shownOpen = open
	a.unpin.guard.Unlock()
	return a.showArranged(full, open)
}

// showArranged puts the window at full; else at the tab cut from it (FR-614).
func (a *App) showArranged(full application.Arrangement, open bool) error {
	if open {
		a.report("giving the ribbon its frame back", a.tabFrame(false))
		return a.place(full.At, full.Size)
	}
	tab, err := a.service.Collapsed(full)
	if err != nil {
		return err
	}
	a.report("taking the frame off the tab", a.tabFrame(true))
	return a.place(tab.At, tab.Size)
}

// ribbonAt answers where the full ribbon stands: where it was last arranged while the window is its
// tab, so a change made while collapsed is fitted from the ribbon's place rather than the tab's.
func (a *App) ribbonAt() (placement.Point, error) {
	a.unpin.guard.Lock()
	collapsed, at := !a.unpin.shownOpen, a.unpin.full.At
	a.unpin.guard.Unlock()
	if collapsed {
		return at, nil
	}
	return a.position()
}

// pointerMoved hears the pointer come onto the ribbon or go off it (FR-615, FR-616).
func (a *App) pointerMoved(arrived bool) {
	if a.pinned() {
		return
	}
	a.changeHover(func(state hover.State, now time.Time) hover.State {
		if arrived {
			return state.Arrived(now)
		}
		return state.Left(now)
	})
}

// hold keeps the ribbon as it is until the matching release (FR-616); holds nest.
func (a *App) hold(expand bool) {
	a.changeHover(func(state hover.State, now time.Time) hover.State {
		a.unpin.holds++
		if expand {
			return state.Held().Expanded(now)
		}
		return state.Held()
	})
}

// release ends one hold; the last one lets the ribbon collapse again.
func (a *App) release() {
	a.changeHover(func(state hover.State, now time.Time) hover.State {
		if a.unpin.holds == 0 {
			return state
		}
		a.unpin.holds--
		if a.unpin.holds > 0 {
			return state
		}
		return state.Released(now)
	})
}

// menuShown holds the ribbon while its own menu is open; menuClosed releases it (FR-616).
func (a *App) menuShown() {
	a.unpin.guard.Lock()
	already := a.unpin.menuHeld
	a.unpin.menuHeld = true
	a.unpin.guard.Unlock()
	if !already {
		a.hold(false)
	}
}

func (a *App) menuClosed() {
	a.unpin.guard.Lock()
	held := a.unpin.menuHeld
	a.unpin.menuHeld = false
	a.unpin.guard.Unlock()
	if held {
		a.release()
	}
}

// changeHover applies change to the hover state at the present moment, schedules the timer for what
// it now awaits, then opens or collapses the window where the state says so. While a panel stands the
// window is that panel, so the ribbon takes its form as the panel closes instead.
func (a *App) changeHover(change func(hover.State, time.Time) hover.State) {
	a.unpin.guard.Lock()
	now := a.now()
	a.unpin.state = change(a.unpin.state, now)
	if a.unpin.stop != nil {
		a.unpin.stop()
		a.unpin.stop = nil
	}
	if due, pending := a.unpin.state.Due(); pending {
		a.unpin.stop = a.after(due.Sub(now), a.hoverDue)
	}
	open := a.unpin.state.Open() || a.pinned()
	changed := open != a.unpin.shownOpen && !a.panelOpen.Load()
	if changed {
		a.unpin.shownOpen = open
	}
	full := a.unpin.full
	a.unpin.guard.Unlock()
	if changed {
		a.report("opening or collapsing the ribbon", a.showArranged(full, open))
		a.emit(eventRefresh)
	}
}

// hoverDue makes the change that has fallen due. It runs on the timer's own goroutine, so a panic is
// caught here and logged rather than ending the application.
func (a *App) hoverDue() {
	defer func() {
		if failure := recover(); failure != nil {
			fmt.Fprintf(a.log, "recovered from %v while opening or collapsing the ribbon\n", failure)
		}
	}()
	a.changeHover(func(state hover.State, now time.Time) hover.State { return state.At(now) })
}

// setPinned pins or unpins the ribbon (FR-613). Pinned, it is shown in full with nothing watching the
// pointer; unpinned, it stays on top (FR-617) and collapses once the pointer has been away for
// hover.Away.
func (a *App) setPinned(on bool) error {
	err := a.service.SetPinned(on)
	a.applyAlwaysOnTop()
	pinned := a.pinned()
	a.unpin.guard.Lock()
	if pinned {
		a.unpin.state = hover.State{}
	} else {
		a.unpin.state = hover.Unpinned(a.now(), false)
	}
	a.unpin.guard.Unlock()
	a.changeHover(func(state hover.State, _ time.Time) hover.State { return state })
	a.trackPointer(!pinned && a.visible.Load())
	return err
}

// trackPointer starts or stops the desktop reporting the pointer, which is wanted only while an
// unpinned ribbon is shown.
func (a *App) trackPointer(on bool) {
	if a.ribbon != 0 {
		a.watchPointer(on)
	}
}
