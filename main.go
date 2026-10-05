// Command TimeRibbon shows a ribbon of clocks, one per chosen place in the world.
//
// This file is the composition root, the only file permitted to import both the application layer
// and concrete infrastructure (TestCompositionRootIsWhitelisted). The window reaches the desktop
// through the shell.Desktop port it is handed here.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/oernster/timeribbon/internal/application"
	"github.com/oernster/timeribbon/internal/infrastructure/store"
	"github.com/oernster/timeribbon/internal/infrastructure/zones"
	"github.com/oernster/timeribbon/internal/product"
	"github.com/oernster/timeribbon/ribbonkit/application/menus"
	"github.com/oernster/timeribbon/ribbonkit/application/release"
	"github.com/oernster/timeribbon/ribbonkit/domain/placement"
	"github.com/oernster/timeribbon/ribbonkit/infrastructure/appdata"
	"github.com/oernster/timeribbon/ribbonkit/infrastructure/desktop"
	"github.com/oernster/timeribbon/ribbonkit/infrastructure/monitors"
	"github.com/oernster/timeribbon/ribbonkit/infrastructure/runlog"
	"github.com/oernster/timeribbon/ribbonkit/infrastructure/startup"
	"github.com/oernster/timeribbon/ribbonkit/infrastructure/system"
	"github.com/oernster/timeribbon/ribbonkit/infrastructure/update"
	"github.com/oernster/timeribbon/ribbonkit/ui/window"
)

// The empty ribbon's one cell, the padding round the cells and the pull out handle's lane, in DIP,
// the same at either size: the prompt holds the large Add clock button whatever size the clocks are
// drawn at; the handle is the same control at both.
var (
	promptCell    = placement.Size{Width: 176, Height: 184}
	ribbonPadding = 6
	handleLane    = 16
)

// layouts is the size of one cell in each style at each size (FR-610), in DIP: its one home. They
// are the least a clock cell is drawn at; a cell is widened to its widest time and date as the
// page measures them (FR-620) and the whole ribbon drawn at the chosen scale (FR-623). app.css
// sizes their text to fit, under .ribbon.small for the small ones.
var layouts = application.Layouts{
	Large: application.Layout{
		Digital:    placement.Size{Width: 176, Height: 92},
		Analogue:   placement.Size{Width: 176, Height: 176},
		Prompt:     promptCell,
		Padding:    ribbonPadding,
		HandleLane: handleLane,
	},
	Small: application.Layout{
		Digital:    placement.Size{Width: 146, Height: 72},
		Analogue:   placement.Size{Width: 146, Height: 116},
		Prompt:     promptCell,
		Padding:    ribbonPadding,
		HandleLane: handleLane,
	},
}

// panels are the window's sizes in DIP while it shows a panel (CON-6): Settings wide enough for its
// choices to sit side by side (FR-625); About, Licence and the update panel narrower, for their text.
var panels = window.PanelSizes{
	Settings: placement.Size{Width: 900, Height: 760},
	Other:    placement.Size{Width: 560, Height: 760},
}

func main() {
	if generatingBindings {
		app, control := newApp(nil, window.Config{Log: io.Discard, Panels: panels})
		if err := control.Run(app, assets, ""); err != nil {
			os.Exit(1)
		}
		return
	}
	log := keepLog()
	if err := run(log); err != nil {
		fmt.Fprintf(log, "%s: %v\n", product.Name, err)
		os.Exit(1)
	}
}

// keepLog opens the log and points standard error at it before anything can fail (NFR-O-1). A log
// that cannot be opened leaves standard error as it is; the run goes on.
func keepLog() io.Writer {
	dir, err := settingsDir()
	if err != nil {
		return os.Stderr
	}
	log, err := runlog.Open(dir, product.App(), time.Now())
	if err != nil {
		return os.Stderr
	}
	if err := runlog.Keep(log); err != nil {
		fmt.Fprintln(log, err)
	}
	return log
}

// settingsDir answers the settings folder appdata names; a folder of the same name in the temporary
// folder when the environment names none, so the ribbon still opens.
func settingsDir() (string, error) {
	dir, err := appdata.Dir(product.App(), os.LookupEnv)
	if err != nil {
		return filepath.Join(os.TempDir(), product.Name), err
	}
	return dir, nil
}

// run wires everything together and hands the facade to Wails. Only a failure to run the window at
// all ends it; every other fault is carried to the ribbon or the log (FR-704, FR-707).
func run(log io.Writer) error {
	dir, err := settingsDir()
	if err != nil {
		fmt.Fprintf(log, "%v: keeping settings in %s\n", err, dir)
	}
	zoneCatalogue, err := zones.New()
	if err != nil {
		return err
	}
	program, err := os.Executable()
	if err != nil {
		fmt.Fprintf(log, "finding this program's path: %v\n", err)
	}
	service := application.New(application.Ports{
		Store:    store.New(dir),
		Zones:    zoneCatalogue,
		Clock:    system.Clock{},
		IDs:      system.IDs{},
		Monitors: monitors.Monitors{},
		Startup:  startup.New(product.App(), program),
		Releases: update.New(),
		Build:    release.Build{Version: product.Version, Platform: release.PlatformKeyFor(runtime.GOOS)},
	}, layouts)
	if err := service.Start(); err != nil {
		fmt.Fprintf(log, "loading settings: %v\n", err)
	}
	var control *window.Control
	desk := desktop.New(product.App(), func() []menus.Item { return service.TrayMenu(control.Visible()) }, log)
	app, control := newApp(service, window.Config{Service: kitService{service}, Desktop: desk, Log: log, Panels: panels})
	preparePlatform(control, desk)
	if err := desk.Start(); err != nil {
		fmt.Fprintf(log, "starting the tray icon: %v; closing the ribbon will exit\n", err)
	} else {
		control.TrayStarted()
	}
	return control.Run(app, assets, dir)
}
