package product

import (
	"runtime"
	"slices"
	"testing"
)

// modulesOf answers the modules credits name.
func modulesOf(credits []Credit) []string {
	var out []string
	for _, credit := range credits {
		out = append(out, credit.Module)
	}
	return out
}

// FR-607: each platform's About names what that platform's build ships; a component with no
// platforms named ships on every one.
func TestEachPlatformCreditsWhatItShips(t *testing.T) {
	t.Parallel()
	windows, linux, mac := modulesOf(CreditsFor(Windows)), modulesOf(CreditsFor(Linux)), modulesOf(CreditsFor(MacOS))
	if !slices.Contains(windows, "github.com/wailsapp/go-webview2") || slices.Contains(linux, "github.com/wailsapp/go-webview2") {
		t.Error("the Windows web view is not credited on Windows alone")
	}
	if !slices.Contains(linux, "github.com/godbus/dbus/v5") || slices.Contains(mac, "github.com/godbus/dbus/v5") {
		t.Error("the Linux tray's bus is not credited on Linux alone")
	}
	for _, list := range [][]string{windows, linux, mac} {
		if !slices.Contains(list, "github.com/wailsapp/wails/v2") {
			t.Error("Wails is not credited on every platform")
		}
	}
	if len(Credits) != len(CreditsFor(runtime.GOOS)) {
		t.Error("Credits is not this platform's list")
	}
}
