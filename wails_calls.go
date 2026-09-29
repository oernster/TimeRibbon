package main

// The facade's production calls into Wails. App holds each as a field so the facade's tests can
// stand in for Wails; newApp points the fields here.

import (
	"math"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// emitToWails sends an event to the page. Before startup there is no context to send it into, so an
// event raised then is dropped.
func (a *App) emitToWails(event string, data ...any) {
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, event, data...)
	}
}

func (a *App) showInWails() { runtime.WindowShow(a.ctx) }

func (a *App) hideInWails() { runtime.WindowHide(a.ctx) }

func (a *App) quitWails() { runtime.Quit(a.ctx) }

func (a *App) setOnTopInWails(on bool) { runtime.WindowSetAlwaysOnTop(a.ctx, on) }

// backgroundInWails paints the window and its web view with an opaque colour.
func (a *App) backgroundInWails(red, green, blue uint8) {
	runtime.WindowSetBackgroundColour(a.ctx, red, green, blue, math.MaxUint8)
}
