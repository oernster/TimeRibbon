package product

import (
	"runtime"
	"slices"
	"strings"
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

// FR-912: every platform's About credits both map pictures to NASA Earth Observatory, saying NASA
// does not endorse TimeRibbon.
func TestEveryPlatformCreditsTheMapPictures(t *testing.T) {
	t.Parallel()
	for _, goos := range Platforms {
		pictures := 0
		for _, credit := range CreditsFor(goos) {
			if credit.Licence == nasaImagery && strings.Contains(credit.Name, "NASA Earth Observatory") && strings.Contains(credit.Licence, "endorsement") {
				pictures++
			}
		}
		if pictures != 2 {
			t.Errorf("%s credits %d map pictures, want 2", goos, pictures)
		}
	}
}
