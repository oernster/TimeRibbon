package main

import (
	"embed"
	"fmt"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"github.com/oernster/timestrip/internal/product"
)

//go:embed all:frontend/dist
var assets embed.FS

// stripClass is the class the strip's window is created with, so it can be found by it (CON-7).
const stripClass = product.Name + "Strip"

// instanceID names the lock that keeps one TimeStrip per user (FR-506).
const instanceID = "uk.codecrafter." + product.Name

// launch runs the window. It starts hidden: startup places it and takes it off the taskbar before
// the page shows it (FR-101).
func launch(app *App) error {
	err := wails.Run(&options.App{
		Title:         product.Name,
		Frameless:     true,
		DisableResize: true,
		StartHidden:   true,
		AssetServer:   &assetserver.Options{Assets: assets},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId:               instanceID,
			OnSecondInstanceLaunch: func(options.SecondInstanceData) { app.secondInstance() },
		},
		Windows: &windows.Options{
			WindowClassName:      stripClass,
			Theme:                windows.SystemDefault,
			DisablePinchZoom:     true,
			IsZoomControlEnabled: false,
		},
		OnStartup:     app.startup,
		OnDomReady:    app.domReady,
		OnBeforeClose: app.beforeClose,
		OnShutdown:    app.shutdown,
		Bind:          []any{app},
	})
	if err != nil {
		return fmt.Errorf("running the window: %w", err)
	}
	return nil
}
