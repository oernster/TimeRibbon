//go:build windows

// Command installer is TimeRibbon's setup program (FR-801 to FR-810).
//
// It is a second Wails application in the module, carrying the built application as an embedded
// payload. It installs, updates, goes back a version, repairs, reinstalls and uninstalls, all per
// user with no administrator rights. The install policy lives in internal/infrastructure/setup;
// this is the window over it.
package main

import (
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	windowsoptions "github.com/wailsapp/wails/v2/pkg/options/windows"
	"golang.org/x/sys/windows"

	"github.com/oernster/timeribbon/internal/infrastructure/setup"
	"github.com/oernster/timeribbon/internal/product"
)

//go:embed all:frontend/dist
var assets embed.FS

// sheet is the page's style sheet, read here only for the two ground colours it states.
//
//go:embed frontend/dist/setup.css
var sheet string

// payload is the built application as a zip archive, embedded as a string so Go keeps it in the
// read-only image rather than charging it to the process. build.ps1 writes the real one before it
// builds and puts the empty placeholder back after.
//
//go:embed payload.zip
var payload string

const (
	// setupID names what is setup's own: its window class, its web view cache and its step log.
	setupID = product.SetupName
	// windowTitle is the setup window's title.
	windowTitle = product.Name + " Setup"
	// windowWidth and windowHeight fit the tallest screen at the sheet's type sizes (DIP). Measured
	// on 2026-09-27 by laying the page out at 804 by 661, this size less a window frame: the tallest
	// screen, Installed, took 341 of the body's 414 pixels and no screen overflowed.
	windowWidth  = 820
	windowHeight = 700
	// opaque is a colour's alpha where nothing shows through it.
	opaque = 255
)

func main() {
	log, logErr := setup.OpenStepLog(filepath.Join(os.TempDir(), setupID+".log"))
	log.Record("setup " + product.Version + " started")
	places, placesErr := setup.ResolvePlaces(os.LookupEnv, windows.KnownFolderPath)
	self, selfErr := os.Executable()
	problem := errors.Join(placesErr, selfErr)
	if problem != nil {
		log.Record("reading the machine: " + problem.Error())
	}
	app := NewApp(Config{
		Machine:     setup.NewMachine(places, setup.AppsList(), setup.StartWithWindows, setup.DeleteAfterExit),
		Processes:   setup.AppProcesses(),
		Log:         log,
		Carried:     setup.Carried{Payload: payload, Self: self, Version: product.Version},
		Args:        os.Args[1:],
		PrefersDark: setup.SystemPrefersDark(),
		Problem:     problem,
	})
	if err := run(app); err != nil {
		log.Record(err.Error())
		if logErr != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}

// run shows the window, painted the page's own ground before the page loads so it never flashes
// the wrong one.
func run(app *App) error {
	light, dark, err := setup.Surfaces(sheet)
	if err != nil {
		return err
	}
	surface := light
	if app.prefersDark {
		surface = dark
	}
	err = wails.Run(&options.App{
		Title:            windowTitle,
		Width:            windowWidth,
		Height:           windowHeight,
		DisableResize:    true,
		BackgroundColour: &options.RGBA{R: surface.R, G: surface.G, B: surface.B, A: opaque},
		AssetServer:      &assetserver.Options{Assets: assets},
		OnStartup:        app.startup,
		OnDomReady:       app.domReady,
		Bind:             []any{app},
		Windows: &windowsoptions.Options{
			WindowClassName: setupID,
			// Setup's own web view cache sits under TEMP, so running it leaves no folder beside the
			// application's settings.
			WebviewUserDataPath: filepath.Join(os.TempDir(), setupID),
		},
	})
	if err != nil {
		return fmt.Errorf("running the setup window: %w", err)
	}
	return nil
}
