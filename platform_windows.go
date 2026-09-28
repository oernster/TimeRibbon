package main

import "github.com/oernster/timeribbon/internal/infrastructure/desktop"

// preparePlatform needs to do nothing on Windows, where the tray reads the icon built into the
// executable.
func preparePlatform(*App, *desktop.Desktop) {}
