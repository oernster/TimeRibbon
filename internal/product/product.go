// Package product holds the product's name: the one home for it, read by the window, the tray,
// the settings folder, the log, the Start with Windows value and the setup program.
package product

// Name is the product's name as a reader sees it.
const Name = "TimeRibbon"

// SetupName is the setup program's name: its executable, its window class, its web view cache and
// its step log.
const SetupName = Name + "Setup"

// RibbonClass is the class the ribbon's window is created with, so it can be found by it (CON-7).
// Setup looks for it to know the ribbon is up before it closes.
const RibbonClass = Name + "Window"

// DonateURL is where the donate button at the foot of Settings sends a browser: the only address
// TimeRibbon knows. It is handed to the desktop to open rather than fetched, so the application
// itself still makes no request (NFR-S-1).
const DonateURL = "https://www.paypal.com/ncp/payment/THUS4KZ5GECH8"

// Version is the version this build carries. build.ps1 stamps it from VERSION into both
// executables with -ldflags -X, which reaches only a var, never a const (CON-4). A build made any
// other way says so by carrying this placeholder.
var Version = "0.0.0-dev"
