package main

import (
	_ "embed"
	"os"
	"os/signal"
	"syscall"

	"github.com/oernster/timeribbon/internal/infrastructure/desktop"
	"github.com/oernster/timeribbon/internal/infrastructure/gtkmain"
)

// trayIcon is the application's icon, which the tray on Linux is handed as an image; on Windows the
// tray reads the icon built into the executable instead.
//
//go:embed build/appicon.png
var trayIcon []byte

// On Linux GTK is sent through X11 before Wails opens it: the ribbon must choose where it stands,
// which a window on Wayland may not.
func init() { gtkmain.ForceX11() }

// preparePlatform gives the desktop the icon, then has a request to end from outside end the
// application.
func preparePlatform(app *App, desk *desktop.Desktop) {
	desk.UseIcon(trayIcon)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	go app.exitWhen(signals)
}
