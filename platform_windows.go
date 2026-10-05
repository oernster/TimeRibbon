package main

import (
	"github.com/oernster/timeribbon/ribbonkit/infrastructure/desktop"
	"github.com/oernster/timeribbon/ribbonkit/ui/window"
)

// preparePlatform needs to do nothing on Windows, where the tray reads the icon built into the
// executable.
func preparePlatform(*window.Control, *desktop.Desktop) {}
