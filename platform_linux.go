package main

import "github.com/oernster/timeribbon/internal/infrastructure/gtkmain"

// On Linux GTK is sent through X11 before Wails opens it: the ribbon must choose where it stands,
// which a window on Wayland may not.
func init() { gtkmain.ForceX11() }
