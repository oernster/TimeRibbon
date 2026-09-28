//go:build windows

package main

import (
	"context"
	"slices"
	"time"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/oernster/timeribbon/internal/infrastructure/setup"
	"github.com/oernster/timeribbon/internal/product"
)

const (
	// progressEvent carries each step's progress to the page.
	progressEvent = "progress"
	// launchWait bounds how long setup waits for the ribbon to show before it closes.
	launchWait = 5 * time.Second
)

// Config is what the setup window is built over.
type Config struct {
	Machine     setup.Machine
	Processes   setup.Processes
	Log         *setup.StepLog
	Carried     setup.Carried
	Args        []string
	PrefersDark bool
	// Problem is why the machine could not be read at all; nil when it could.
	Problem error
}

// App is the Wails facade: everything the page can do goes through a method here. Each one hands
// straight to the setup package, which owns the install policy.
type App struct {
	ctx         context.Context
	machine     setup.Machine
	processes   setup.Processes
	log         *setup.StepLog
	carried     setup.Carried
	uninstall   bool
	prefersDark bool
	problem     error
}

// NewApp builds the facade. Started with -uninstall, setup opens on the Uninstall screen (FR-801).
func NewApp(config Config) *App {
	return &App{
		machine:     config.Machine,
		processes:   config.Processes,
		log:         config.Log,
		carried:     config.Carried,
		uninstall:   slices.Contains(config.Args, setup.UninstallFlag),
		prefersDark: config.PrefersDark,
		problem:     config.Problem,
	}
}

func (a *App) startup(ctx context.Context) { a.ctx = ctx }

// domReady gives the page the keyboard once it exists, since a cold launch can lose the race that
// would otherwise hand it over.
func (a *App) domReady(context.Context) { a.TakeKeyboard() }

// TakeKeyboard gives the web view the keyboard; the page calls it when it finds it has none.
func (a *App) TakeKeyboard() {
	if !setup.TakeFocus(setupID) && a.ctx != nil {
		wailsruntime.WindowShow(a.ctx)
	}
}

// StateDTO is the one reading of the machine the page routes on. AppName travels with it because
// the page must not write the product's name down.
type StateDTO struct {
	AppName          string `json:"appName"`
	Route            string `json:"route"`
	Uninstall        bool   `json:"uninstall"`
	InstalledVersion string `json:"installedVersion"`
	ThisVersion      string `json:"thisVersion"`
	StartMenu        bool   `json:"startMenu"`
	Desktop          bool   `json:"desktop"`
	StartWithWindows bool   `json:"startWithWindows"`
	PrefersDark      bool   `json:"prefersDark"`
	LogPath          string `json:"logPath"`
	// Problem is why the machine could not be read; the page shows it as the failure verdict.
	Problem string `json:"problem"`
}

// ChoicesDTO carries the three boxes of FR-805 from the page.
type ChoicesDTO struct {
	StartMenu        bool `json:"startMenu"`
	Desktop          bool `json:"desktop"`
	StartWithWindows bool `json:"startWithWindows"`
}

func (c ChoicesDTO) choices() setup.Choices {
	return setup.Choices{StartMenu: c.StartMenu, Desktop: c.Desktop, StartWithWindows: c.StartWithWindows}
}

// ProgressDTO is emitted on the progress event while work runs.
type ProgressDTO struct {
	Pct int    `json:"pct"`
	Msg string `json:"msg"`
}

// DetectState reads the machine once and decides the route (FR-801).
func (a *App) DetectState() StateDTO {
	state := StateDTO{
		AppName:     setup.AppName,
		Uninstall:   a.uninstall,
		ThisVersion: a.carried.Version,
		PrefersDark: a.prefersDark,
		LogPath:     a.log.Path(),
	}
	if a.problem != nil {
		state.Problem = a.problem.Error()
		return state
	}
	existing, err := a.machine.Read()
	if err != nil {
		a.log.Record("reading the machine: " + err.Error())
		state.Problem = err.Error()
		return state
	}
	offered := setup.Offered(existing)
	state.Route = string(setup.RouteFor(existing, a.carried.Version))
	state.InstalledVersion = existing.Version
	state.StartMenu, state.Desktop, state.StartWithWindows = offered.StartMenu, offered.Desktop, offered.StartWithWindows
	a.log.Record("route " + state.Route + ", installed " + existing.Version + ", carried " + a.carried.Version)
	return state
}

// AppRunning reports whether TimeRibbon is open, asked before any file is touched (FR-807).
func (a *App) AppRunning() bool { return a.processes.Running() }

// CloseRunningApp ends every running copy by image name and waits for them to go (FR-807).
func (a *App) CloseRunningApp() error {
	a.log.Record("closing the running copy")
	err := a.processes.Close()
	if err != nil {
		a.log.Record(err.Error())
	}
	return err
}

// Install performs an install, an update, a going back or a reinstall: all the same act (FR-802).
func (a *App) Install(choices ChoicesDTO) error {
	return a.perform("install", func() ([]setup.Step, error) {
		return a.machine.InstallSteps(a.carried, choices.choices()), nil
	})
}

// Repair writes the files again, keeping every box as it stands on the machine (FR-804).
func (a *App) Repair() error {
	return a.perform("repair", func() ([]setup.Step, error) { return a.machine.RepairSteps(a.carried) })
}

// Uninstall removes TimeRibbon, forgetting the settings only when asked (FR-806).
func (a *App) Uninstall(forget bool) error {
	return a.perform("uninstall", func() ([]setup.Step, error) { return a.machine.UninstallSteps(forget), nil })
}

// Apply applies the boxes at once, as each changes on the Installed screen.
func (a *App) Apply(choices ChoicesDTO) error {
	a.log.Record("applying the boxes")
	return setup.Run(a.machine.ChoiceSteps(choices.choices()), a.log, func(setup.Progress) {})
}

// perform checks nothing is running, then runs the steps with the bar and the step log.
func (a *App) perform(name string, steps func() ([]setup.Step, error)) error {
	a.log.Record(name + " asked for")
	if a.processes.Running() {
		a.log.Record(setup.ErrAppRunning.Error())
		return setup.ErrAppRunning
	}
	list, err := steps()
	if err != nil {
		a.log.Record(err.Error())
		return err
	}
	return setup.Run(list, a.log, a.progress)
}

// LaunchApp starts TimeRibbon and waits for the ribbon to come forward, so setup closes behind it.
func (a *App) LaunchApp() error {
	a.log.Record("starting " + setup.AppName)
	err := setup.Launch(a.machine.Places().Program(), product.RibbonClass, launchWait)
	if err != nil {
		a.log.Record(err.Error())
	}
	return err
}

// Licence answers the licence setup carries, for the Licence screen.
func (a *App) Licence() (string, error) { return setup.Licence(a.carried.Payload) }

// Quit closes the setup program.
func (a *App) Quit() {
	a.log.Record("closed")
	wailsruntime.Quit(a.ctx)
}

// progress reports how far the work has got.
func (a *App) progress(p setup.Progress) {
	wailsruntime.EventsEmit(a.ctx, progressEvent, ProgressDTO{Pct: p.Percent, Msg: p.Step})
}
