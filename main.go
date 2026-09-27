// Command TimeStrip shows a strip of clocks, one per chosen place in the world.
//
// This file is the composition root. It and app.go are the only files permitted to wire concrete
// infrastructure to the application layer.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/oernster/timestrip/internal/application"
	"github.com/oernster/timestrip/internal/domain/placement"
	"github.com/oernster/timestrip/internal/infrastructure/appdata"
	"github.com/oernster/timestrip/internal/infrastructure/desktop"
	"github.com/oernster/timestrip/internal/infrastructure/monitors"
	"github.com/oernster/timestrip/internal/infrastructure/runlog"
	"github.com/oernster/timestrip/internal/infrastructure/startup"
	"github.com/oernster/timestrip/internal/infrastructure/store"
	"github.com/oernster/timestrip/internal/infrastructure/system"
	"github.com/oernster/timestrip/internal/infrastructure/zones"
	"github.com/oernster/timestrip/internal/product"
)

// layout is the size of one cell in each style and the padding round the cells, in DIP: its one
// home. The page draws cells at these sizes from the snapshot.
var layout = application.Layout{
	Digital:  placement.Size{Width: 176, Height: 92},
	Analogue: placement.Size{Width: 176, Height: 176},
	Prompt:   placement.Size{Width: 176, Height: 184},
	Padding:  6,
}

// panelSize is the window in DIP while it shows a panel: Settings, About or Licence (CON-6).
var panelSize = placement.Size{Width: 560, Height: 760}

func main() {
	if generatingBindings {
		if err := launch(&App{}); err != nil {
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
	log, err := runlog.Open(dir, time.Now())
	if err != nil {
		return os.Stderr
	}
	if err := runlog.Keep(log); err != nil {
		fmt.Fprintln(log, err)
	}
	return log
}

// settingsDir answers %APPDATA%\TimeStrip; a folder of the same name in the temporary folder when
// the environment names none, so the strip still opens.
func settingsDir() (string, error) {
	dir, err := appdata.Dir(os.LookupEnv)
	if err != nil {
		return filepath.Join(os.TempDir(), product.Name), err
	}
	return dir, nil
}

// run wires everything together and hands the facade to Wails. Only a failure to run the window at
// all ends it; every other fault is carried to the strip or the log (FR-704, FR-707).
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
		Startup:  startup.New(program),
	}, layout)
	if err := service.Start(); err != nil {
		fmt.Fprintf(log, "loading settings: %v\n", err)
	}
	var app *App
	desk := desktop.New(func() []application.MenuItem { return service.TrayMenu(app.visible.Load()) }, log)
	app = newApp(service, desk, log, panelSize)
	if err := desk.Start(); err != nil {
		fmt.Fprintf(log, "starting the tray icon: %v; closing the strip will exit\n", err)
	} else {
		app.trayUp.Store(true)
	}
	return launch(app)
}
