//go:build linux || darwin

package main

import (
	_ "embed"
	"os"
	"os/signal"
	"syscall"

	"github.com/oernster/timeribbon/ribbonkit/infrastructure/desktop"
	"github.com/oernster/timeribbon/ribbonkit/ui/window"
)

// trayIcon is the application's icon, which the tray on Linux and the menu bar on macOS are handed
// as an image; on Windows the tray reads the icon built into the executable instead.
//
//go:embed build/appicon.png
var trayIcon []byte

// preparePlatform gives the desktop the icon, then has a request to end from outside end the
// application.
func preparePlatform(control *window.Control, desk *desktop.Desktop) {
	desk.UseIcon(trayIcon)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	go control.ExitWhen(signals)
}
