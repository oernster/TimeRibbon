//go:build windows

// Command installer is TimeRibbon's setup program (FR-801 to FR-810).
//
// It is a second Wails application in the module, carrying the built application as an embedded
// payload. It installs, updates, goes back a version, repairs, reinstalls and uninstalls, all per
// user with no administrator rights. The window, its page and the install policy are ribbonkit's
// (ribbonkit/installer over ribbonkit/infrastructure/setup), which name no product; this is the
// composition root that names it and carries what is TimeRibbon's own.
package main

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows"

	"github.com/oernster/timeribbon/internal/product"
	"github.com/oernster/timeribbon/ribbonkit/infrastructure/setup"
	"github.com/oernster/timeribbon/ribbonkit/installer"
)

// pictures are the page's pictures, which tools/genicons.py makes from TimeRibbon's artwork: the
// header mark and the theme switch's sun and moon.
//
//go:embed all:frontend/dist
var pictures embed.FS

// picturesRoot is the folder pictures holds them under.
const picturesRoot = "frontend/dist"

// payload is the built application as a zip archive, embedded as a string so Go keeps it in the
// read-only image rather than charging it to the process. build.ps1 writes the real one before it
// builds and puts the empty placeholder back after.
//
//go:embed payload.zip
var payload string

// setupID names what is setup's own: its window class, its web view cache and its step log.
const setupID = product.SetupName

// App is what Wails binds, so the page reaches the kit's setup facade as main.App.
type App struct{ *installer.Setup }

func main() {
	log, logErr := setup.OpenStepLog(filepath.Join(os.TempDir(), setupID+".log"))
	log.Record("setup " + product.Version + " started")
	installs := setup.Product{App: product.App(), Publisher: product.Author}
	places, placesErr := setup.ResolvePlaces(installs, os.LookupEnv, windows.KnownFolderPath)
	self, selfErr := os.Executable()
	problem := errors.Join(placesErr, selfErr)
	if problem != nil {
		log.Record("reading the machine: " + problem.Error())
	}
	facade := installer.New(installer.Config{
		Product:     installs,
		SetupID:     setupID,
		RibbonClass: product.RibbonClass,
		Machine:     setup.NewMachine(installs, places, setup.AppsList(installs), setup.StartWithWindows(installs.App), setup.DeleteAfterExit),
		Processes:   setup.AppProcesses(installs),
		Log:         log,
		Carried:     setup.Carried{Payload: payload, Self: self, Version: product.Version},
		Args:        os.Args[1:],
		PrefersDark: setup.SystemPrefersDark(),
		Problem:     problem,
	})
	if err := run(facade); err != nil {
		log.Record(err.Error())
		if logErr != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}

// run shows the setup window with TimeRibbon's title and pictures.
func run(facade *installer.Setup) error {
	shown, err := fs.Sub(pictures, picturesRoot)
	if err != nil {
		return fmt.Errorf("reading the setup pictures: %w", err)
	}
	return installer.Run(&App{facade}, facade, installer.Window{Title: product.Name + " Setup", Pictures: shown})
}
