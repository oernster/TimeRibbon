package main

import (
	"embed"
	"fmt"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"github.com/oernster/timeribbon/internal/product"
)

//go:embed all:frontend/dist
var assets embed.FS

// instanceID names the lock that keeps one TimeRibbon per user (FR-506).
const instanceID = product.AppID

// webViewFolder names the web view's own data folder inside the settings folder. Left to Wails, it
// would be a folder named for the executable beside the settings folder, which uninstalling with
// "Also forget my settings" did not reach (FR-806); inside it, forgetting removes it with the rest.
const webViewFolder = "WebView2"

// launch runs the window, keeping the web view's data in the settings folder dir; where dir is
// empty, as in the run that generates bindings, Wails chooses. It starts hidden: startup places it
// and takes it off the taskbar before the page shows it (FR-101).
func launch(app *App, dir string) error {
	webViewData := ""
	if dir != "" {
		webViewData = filepath.Join(dir, webViewFolder)
	}
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
		// The web view is transparent, so wherever the page draws less than opaque the desktop shows
		// through; the window's own paint follows the chosen opacity (FR-622, opacity.go).
		Windows: &windows.Options{
			WebviewIsTransparent: true,
			WindowClassName:      product.RibbonClass,
			WebviewUserDataPath:  webViewData,
			Theme:                windows.SystemDefault,
			DisablePinchZoom:     true,
			IsZoomControlEnabled: false,
		},
		Mac:           &mac.Options{WebviewIsTransparent: true},
		Linux:         &linux.Options{WindowIsTranslucent: true},
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
