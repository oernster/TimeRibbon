package store

// What every ribbon's settings file does alike (a file kept aside, never saved over when it could not
// be read, unknown keys kept, a byte order mark, ids told apart, ordering by position, atomic writes)
// is proved once, in the kit's settingsfile. These tests prove what TimeRibbon's file holds.

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/oernster/ribbonkit/domain/localtime"
	"github.com/oernster/ribbonkit/domain/placement"
	"github.com/oernster/ribbonkit/domain/ribbon"
	"github.com/oernster/ribbonkit/infrastructure/settingsfile"
	"github.com/oernster/timeribbon/internal/domain/clock"
	"github.com/oernster/timeribbon/internal/domain/settings"
	"github.com/oernster/timeribbon/internal/product"
)

// write puts text in dir's settings file.
func write(t *testing.T, dir, text string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, settingsfile.FileName), []byte(text), settingsfile.FileMode); err != nil {
		t.Fatal(err)
	}
}

// read answers dir's settings file.
func read(t *testing.T, dir string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(dir, settingsfile.FileName))
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func full() settings.Settings {
	s := settings.Settings{
		Choices: ribbon.Choices{
			Colour: ribbon.Sunset, Orientation: ribbon.Vertical, Theme: ribbon.Dark, AlwaysOnTop: true,
			SkippedUpdate: "v2.1.0",
			Placement: &placement.Stored{
				Device: `\\.\DISPLAY2`, Work: placement.Rect{Left: 1920, Right: 4480, Bottom: 1392},
				DPI: 144, Offset: placement.Point{X: 180, Y: -4},
			},
			LastEdge: &placement.Against{Device: `\\.\DISPLAY2`, Edge: placement.Left},
			Opacity:  55, Scale: 150,
			PullOutSide: placement.Right,
		},
		Style: settings.Analogue, Size: settings.Small, Format: localtime.TwelveHour, DateFormat: clock.MonthDayYear,
		SunMap: true, PullOut: true,
	}
	s = s.WithClockAdded(settings.Entry{ID: "a1", Zone: "America/New_York", Label: "New York"})
	return s.WithClockAdded(settings.Entry{ID: "b2", Zone: "Australia/Sydney", Label: "Mum"})
}

// FR-701: every value is written and read back, each away from its default.
func TestSettingsRoundTrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := New(dir, product.Name).Save(full()); err != nil {
		t.Fatal(err)
	}
	loaded, err := New(dir, product.Name).Load()
	if err != nil || loaded.Notice != "" {
		t.Fatalf("load: %v %q", err, loaded.Notice)
	}
	got, want := loaded.Settings, full()
	if got.Style != want.Style || got.Size != want.Size || got.Format != want.Format || got.DateFormat != want.DateFormat ||
		got.SunMap != want.SunMap || got.PullOut != want.PullOut || !slices.Equal(got.Clocks, want.Clocks) {
		t.Errorf("got %+v", got)
	}
	if got.Colour != want.Colour || *got.Placement != *want.Placement || *got.LastEdge != *want.LastEdge ||
		got.Opacity != want.Opacity || got.Pinned != want.Pinned || got.PullOutSide != want.PullOutSide {
		t.Errorf("the ribbon's choices read as %+v", got.Choices)
	}
}

// FR-701: the file is indented, in writing order, holding nothing derived.
func TestNoDerivedValueIsStored(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := New(dir, product.Name).Save(full()); err != nil {
		t.Fatal(err)
	}
	text := read(t, dir)
	for _, derived := range []string{"offset\": -", "abbreviation", "EDT", "\"time\"", "\"date\""} {
		if strings.Contains(text, derived) {
			t.Errorf("the file holds %q:\n%s", derived, text)
		}
	}
	if !strings.HasPrefix(text, "{\n  \"version\": 1,\n  \"style\": \"analogue\",") {
		t.Errorf("not indented in writing order:\n%s", text)
	}
	if !strings.Contains(text, `"position": 1`) {
		t.Errorf("positions are not written:\n%s", text)
	}
}

// FR-703.
func TestAbsentFileMeansDefaults(t *testing.T) {
	t.Parallel()
	loaded, err := New(t.TempDir(), product.Name).Load()
	if err != nil || loaded.Notice != "" || loaded.Settings.Style != settings.Digital || len(loaded.Settings.Clocks) != 0 {
		t.Errorf("got %+v (%v)", loaded, err)
	}
}

// FR-704: clocks that are not a list mean nothing in the file can be trusted, so it is kept aside.
func TestClocksThatAreNotAListKeepTheFileAside(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	write(t, dir, `{"clocks": 5}`)
	loaded, err := New(dir, product.Name).Load()
	if err != nil || loaded.Notice != settingsfile.KeptAsideNotice(settingsfile.UnreadableName) || len(loaded.Settings.Clocks) != 0 {
		t.Errorf("got %+v (%v)", loaded, err)
	}
}

// FR-705, FR-706: the acceptance example, plus entries missing an id or a zone.
func TestOneBadClockLeavesTheOthersWorking(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	bad := `{"id": "b", "zone": 7}`
	write(t, dir, `{"clocks": [
		{"id": "a", "zone": "Europe/London", "label": "London", "position": 0},
		`+bad+`,
		{"id": "c", "zone": "Not/AZone", "label": "Gran", "position": 2},
		{"zone": "Asia/Tokyo"},
		{"id": "e", "label": "no zone"}
	]}`)
	store := New(dir, product.Name)
	loaded, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	clocks := loaded.Settings.Clocks
	if len(clocks) != 5 || clocks[0].Label != "London" || clocks[2] != (settings.Entry{ID: "c", Zone: "Not/AZone", Label: "Gran"}) {
		t.Fatalf("clocks %+v", clocks)
	}
	for index, reason := range map[int]string{1: "fields are not what a clock holds", 3: "has no id", 4: "names no time zone"} {
		if !strings.Contains(clocks[index].Unreadable, reason) || !strings.HasPrefix(clocks[index].ID, settingsfile.UnreadableIDPrefix) {
			t.Errorf("entry %d: %+v", index, clocks[index])
		}
	}
	if err := store.Save(loaded.Settings); err != nil {
		t.Fatal(err)
	}
	var file struct{ Clocks []json.RawMessage }
	if err := json.Unmarshal([]byte(read(t, dir)), &file); err != nil || len(file.Clocks) != len(clocks) {
		t.Fatalf("the file reads as %+v (%v)", file, err)
	}
	var written, found bytes.Buffer
	if json.Compact(&written, file.Clocks[1]) != nil || json.Compact(&found, []byte(bad)) != nil || written.String() != found.String() {
		t.Errorf("the bad entry was written back as %s", file.Clocks[1])
	}
}

func TestABadValueLeavesItsDefaultAndTheRestLoad(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	write(t, dir, `{"style": 7, "theme": "dark", "alwaysOnTop": "yes", "placement": {"device": 3},
		"clocks": null}`)
	loaded, err := New(dir, product.Name).Load()
	got := loaded.Settings
	if err != nil || got.Style != settings.Digital || got.Theme != ribbon.Dark || got.AlwaysOnTop || got.Placement != nil {
		t.Errorf("got %+v (%v)", got, err)
	}
}
