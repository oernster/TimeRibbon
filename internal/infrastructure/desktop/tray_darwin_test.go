package desktop

import (
	"errors"
	"io"
	"os"
	"testing"
	"time"

	"github.com/oernster/timeribbon/internal/application"
)

// The icon appears once AppKit's loop has served the request, so its arrival is polled for.
const (
	iconLimit = 2 * time.Second
	iconPause = 20 * time.Millisecond
)

// FR-501: the icon stands in the menu bar with its image until Stop takes it out again.
func TestTheIconStandsInTheMenuBarUntilStop(t *testing.T) {
	icon, err := os.ReadFile("../../../build/appicon.png")
	if err != nil {
		t.Fatal(err)
	}
	d := New(func() []application.MenuItem { return nil }, io.Discard)
	d.UseIcon(icon)
	if err := d.Start(); err != nil {
		t.Fatal(err)
	}
	shown := false
	for deadline := time.Now().Add(iconLimit); time.Now().Before(deadline) && !shown; time.Sleep(iconPause) {
		shown = shownInMenuBar()
	}
	if !shown {
		t.Error("the icon did not appear in the menu bar")
	}
	d.Stop()
	if shownInMenuBar() {
		t.Error("the icon is still in the menu bar after Stop")
	}
}

// A desktop given no image says so rather than showing an empty icon.
func TestAnIconWithNoImageIsRefused(t *testing.T) {
	d := New(func() []application.MenuItem { return nil }, io.Discard)
	defer d.Stop()
	if err := d.Start(); !errors.Is(err, errNoIcon) {
		t.Errorf("got %v", err)
	}
}
