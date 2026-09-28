package desktop

// GTK's side of the Linux desktop: the signals it calls back on and the ribbon's popup menu. The C
// half is in desktop_linux.c, since a Go file that exports to C may only declare C functions.

/*
#cgo pkg-config: gtk+-3.0
#include <stdlib.h>
#include <gtk/gtk.h>
void desktop_watch(GtkWindow *ribbon, guintptr handle);
GtkWidget *menu_new(void);
void menu_add_item(GtkWidget *menu, const char *label, gboolean checkable, gboolean checked, int index);
GtkWidget *menu_add_submenu(GtkWidget *menu, const char *label);
void menu_add_separator(GtkWidget *menu);
void desktop_popup(GtkWindow *ribbon, GtkWidget *menu);
*/
import "C"

import (
	"runtime/cgo"
	"sync"
	"unsafe"

	"github.com/oernster/timeribbon/internal/application"
	"github.com/oernster/timeribbon/internal/domain/placement"
)

// shown is the menu on screen: the desktop its choice goes to and the items it was built from. One
// popup menu is open at a time, as GTK allows.
var shown struct {
	sync.Mutex
	desktop *Desktop
	items   []application.MenuItem
}

// watchWindow connects the ribbon's moves and the screen's changes of display to the desktop that
// handle holds.
func watchWindow(ribbon Window, handle cgo.Handle) error {
	return onWindow(ribbon, func(window *C.GtkWindow) { C.desktop_watch(window, C.guintptr(handle)) })
}

// popUp shows items over ribbon at the pointer, their choice going to d.
func popUp(ribbon Window, d *Desktop, items []application.MenuItem) error {
	shown.Lock()
	shown.desktop, shown.items = d, items
	shown.Unlock()
	return onWindow(ribbon, func(window *C.GtkWindow) {
		menu := C.menu_new()
		next := 0
		build(menu, items, &next)
		C.desktop_popup(window, menu)
	})
}

// build adds items to menu, a submenu for each item holding children. Numbers are given out depth
// first from next, the order actionAt reads them in.
func build(menu *C.GtkWidget, items []application.MenuItem, next *int) {
	for _, item := range items {
		if separatedBefore(item) {
			C.menu_add_separator(menu)
		}
		label := C.CString(item.Label)
		if len(item.Children) > 0 {
			sub := C.menu_add_submenu(menu, label)
			C.free(unsafe.Pointer(label))
			build(sub, item.Children, next)
			continue
		}
		C.menu_add_item(menu, label, gboolean(item.Checkable), gboolean(item.Checked), C.int(*next))
		C.free(unsafe.Pointer(label))
		*next++
	}
}

// gboolean answers GTK's truth value for b.
func gboolean(b bool) C.gboolean {
	if b {
		return C.TRUE
	}
	return C.FALSE
}

//export desktopMoved
func desktopMoved(handle uintptr, x, y C.int) {
	cgo.Handle(handle).Value().(*Desktop).moved(placement.Point{X: int(x), Y: int(y)})
}

//export desktopDisplaysChanged
func desktopDisplaysChanged(handle uintptr) {
	cgo.Handle(handle).Value().(*Desktop).send(Event{Kind: EventDisplayChanged})
}

//export desktopMenuChosen
func desktopMenuChosen(index C.int) {
	shown.Lock()
	d, items := shown.desktop, shown.items
	shown.Unlock()
	if action, ok := actionAt(items, int(index)); ok && d != nil {
		d.send(Event{Kind: EventMenu, Action: action})
	}
}
