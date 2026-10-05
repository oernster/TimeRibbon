package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/oernster/timeribbon/internal/domain/clock"
	"github.com/oernster/timeribbon/internal/domain/settings"
	"github.com/oernster/timeribbon/ribbonkit/domain/placement"
)

// write puts text in dir's settings file.
func write(t *testing.T, dir, text string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, FileName), []byte(text), fileMode); err != nil {
		t.Fatal(err)
	}
}

// read answers dir's settings file.
func read(t *testing.T, dir string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(dir, FileName))
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func full() settings.Settings {
	s := settings.Settings{
		Style: settings.Analogue, Size: settings.Small, Colour: settings.Sunset, Format: clock.TwelveHour, Orientation: settings.Vertical,
		Theme: settings.Dark, AlwaysOnTop: true, SkippedUpdate: "v2.1.0", DateFormat: clock.MonthDayYear,
		Placement: &placement.Stored{
			Device: `\\.\DISPLAY2`, Work: placement.Rect{Left: 1920, Right: 4480, Bottom: 1392},
			DPI: 144, Offset: placement.Point{X: 180, Y: -4},
		},
		LastEdge: &placement.Against{Device: `\\.\DISPLAY2`, Edge: placement.Left},
		SunMap:   true, PullOut: true, Opacity: 55, Scale: 150,
	}
	s = s.WithClockAdded(settings.Entry{ID: "a1", Zone: "America/New_York", Label: "New York"})
	return s.WithClockAdded(settings.Entry{ID: "b2", Zone: "Australia/Sydney", Label: "Mum"})
}

// FR-701.
func TestSettingsRoundTrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := New(dir).Save(full()); err != nil {
		t.Fatal(err)
	}
	loaded, err := New(dir).Load()
	if err != nil || loaded.Notice != "" {
		t.Fatalf("load: %v %q", err, loaded.Notice)
	}
	got, want := loaded.Settings, full()
	if got.Style != want.Style || got.Size != want.Size || got.Colour != want.Colour || got.Format != want.Format || got.Orientation != want.Orientation ||
		got.Theme != want.Theme || got.AlwaysOnTop != want.AlwaysOnTop || got.SkippedUpdate != want.SkippedUpdate || got.DateFormat != want.DateFormat || *got.Placement != *want.Placement ||
		!slices.Equal(got.Clocks, want.Clocks) {
		t.Errorf("got %+v", got)
	}
	// full's opacity is away from the default, so it is proved written and read (FR-622).
	if got.Opacity != want.Opacity || got.Scale != want.Scale {
		t.Errorf("opacity and scale read as %d and %d, want %d and %d", got.Opacity, got.Scale, want.Opacity, want.Scale)
	}
	// full is unpinned, away from the default, so the pin is proved written and read (FR-613).
	if got.Pinned != want.Pinned {
		t.Errorf("pinned read as %v, want %v", got.Pinned, want.Pinned)
	}
	// FR-901, FR-903: full has the sun map on and pulled out, away from the defaults.
	if !got.SunMap || !got.PullOut {
		t.Errorf("sun map read as %v, pull out %v; want both on", got.SunMap, got.PullOut)
	}
	// FR-411: the remembered edge is written and read.
	if got.LastEdge == nil || *got.LastEdge != *want.LastEdge {
		t.Errorf("last edge read as %+v, want %+v", got.LastEdge, want.LastEdge)
	}
}

// FR-411: a remembered edge that is missing, malformed or names no display is none.
func TestAnUnreadableLastEdgeIsNone(t *testing.T) {
	t.Parallel()
	for name, body := range map[string]string{
		"missing":    `{"version": 1}`,
		"not a list": `{"version": 1, "lastEdge": [1]}`,
		"no display": `{"version": 1, "lastEdge": {"edge": "left"}}`,
	} {
		decoded, _, ok := decode([]byte(body))
		if !ok || decoded.LastEdge != nil {
			t.Errorf("%s: read %+v", name, decoded.LastEdge)
		}
	}
}

// FR-701: the file is indented, in writing order, holding nothing derived.
func TestNoDerivedValueIsStored(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := New(dir).Save(full()); err != nil {
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
	dir := filepath.Join(t.TempDir(), "TimeRibbon")
	loaded, err := New(dir).Load()
	if err != nil || loaded.Notice != "" || loaded.Settings.Style != settings.Digital || len(loaded.Settings.Clocks) != 0 {
		t.Errorf("got %+v (%v)", loaded, err)
	}
	if _, err := os.Stat(dir); !errors.Is(err, os.ErrNotExist) {
		t.Error("loading made the folder")
	}
}

// FR-704.
func TestUnreadableFileIsKeptAsideAndReported(t *testing.T) {
	t.Parallel()
	for name, text := range map[string]string{"not JSON": "{ this is not", "clocks not a list": `{"clocks": 5}`, "a list": `[1]`} {
		dir := t.TempDir()
		write(t, dir, text)
		loaded, err := New(dir).Load()
		if err != nil || loaded.Notice != keptAsideNotice(UnreadableName) || len(loaded.Settings.Clocks) != 0 {
			t.Errorf("%s: got %+v (%v)", name, loaded, err)
		}
		kept, err := os.ReadFile(filepath.Join(dir, UnreadableName))
		if err != nil || string(kept) != text {
			t.Errorf("%s: kept %q (%v)", name, kept, err)
		}
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
	store := New(dir)
	loaded, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	clocks := loaded.Settings.Clocks
	if len(clocks) != 5 || clocks[0].Label != "London" || clocks[2] != (settings.Entry{ID: "c", Zone: "Not/AZone", Label: "Gran"}) {
		t.Fatalf("clocks %+v", clocks)
	}
	for index, reason := range map[int]string{1: "fields are not what a clock holds", 3: "has no id", 4: "names no time zone"} {
		if !strings.Contains(clocks[index].Unreadable, reason) || !strings.HasPrefix(clocks[index].ID, unreadableIDPrefix) {
			t.Errorf("entry %d: %+v", index, clocks[index])
		}
	}
	if err := store.Save(loaded.Settings); err != nil {
		t.Fatal(err)
	}
	if text := compact(t, read(t, dir)); !strings.Contains(text, compact(t, bad)) {
		t.Errorf("the bad entry was not written back as found:\n%s", text)
	}
}

// The stored position keeps the order clocks were added in, which settles ties in the ribbon's time
// order (FR-102) and is part of the file's contract (NFR-C-1); an entry with none keeps its place.
func TestOrderingPersistsByPosition(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	write(t, dir, `{"clocks": [
		{"id": "late", "zone": "Asia/Tokyo", "position": 2},
		{"id": "none", "zone": "Europe/Paris"},
		{"id": "early", "zone": "Europe/London", "position": 0}
	]}`)
	loaded, _ := New(dir).Load()
	var ids []string
	for _, entry := range loaded.Settings.Clocks {
		ids = append(ids, entry.ID)
	}
	if !slices.Equal(ids, []string{"early", "none", "late"}) {
		t.Errorf("got %v", ids)
	}
}

func TestABadValueLeavesItsDefaultAndTheRestLoad(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	write(t, dir, `{"style": 7, "theme": "dark", "alwaysOnTop": "yes", "placement": {"device": 3},
		"clocks": null}`)
	loaded, err := New(dir).Load()
	got := loaded.Settings
	if err != nil || got.Style != settings.Digital || got.Theme != settings.Dark || got.AlwaysOnTop || got.Placement != nil {
		t.Errorf("got %+v (%v)", got, err)
	}
}

// Silence check: a later version's keys survive this version saving.
func TestUnknownKeysAreKeptOnWrite(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	write(t, dir, `{"version": 2, "future": {"x": [1, 2]}, "style": "digital", "later": true}`)
	store := New(dir)
	loaded, _ := store.Load()
	if err := store.Save(loaded.Settings); err != nil {
		t.Fatal(err)
	}
	text := read(t, dir)
	var object map[string]any
	if err := json.Unmarshal([]byte(text), &object); err != nil {
		t.Fatal(err)
	}
	if object["later"] != true || compact(t, text) == "" || !strings.Contains(compact(t, text), `"future":{"x":[1,2]}`) {
		t.Errorf("an unknown key was lost:\n%s", text)
	}
	clocks, future, later := strings.Index(text, `"clocks"`), strings.Index(text, `"future"`), strings.Index(text, `"later"`)
	if !(clocks < future && future < later) {
		t.Errorf("unknown keys are not written after the known ones in their order:\n%s", text)
	}
}

// FR-704: when the unreadable file cannot be kept aside, every name being taken, nothing
// overwrites it.
func TestAFileThatCannotBeKeptAsideIsNeverOverwritten(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	write(t, dir, "{ not JSON")
	for n := 1; n <= keptAsideLimit; n++ {
		if err := os.Mkdir(filepath.Join(dir, keptAsideName(n)), folderMode); err != nil {
			t.Fatal(err)
		}
	}
	store := New(dir)
	if _, err := store.Load(); !errors.Is(err, ErrNotKeptAside) {
		t.Errorf("load answered %v", err)
	}
	if err := store.Save(settings.Defaults()); !errors.Is(err, ErrNotKeptAside) {
		t.Errorf("save answered %v", err)
	}
	if read(t, dir) != "{ not JSON" {
		t.Error("the unreadable file was overwritten")
	}
}

// FR-702: a replacement that fails leaves no temporary file behind.
func TestAFailedReplacementLeavesNoTemporaryFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, FileName, "inside"), folderMode); err != nil {
		t.Fatal(err)
	}
	if err := New(dir).Save(settings.Defaults()); err == nil {
		t.Error("replacing a folder succeeded")
	}
	matches, _ := filepath.Glob(filepath.Join(dir, tempPattern))
	if len(matches) != 0 {
		t.Errorf("left behind %v", matches)
	}
}

// FR-702: a save replaces the file whole and leaves no temporary file behind.
func TestWriteReplacesAtomically(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	write(t, dir, `{"style": "analogue"}`)
	if err := New(dir).Save(settings.Defaults()); err != nil {
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 || entries[0].Name() != FileName {
		t.Errorf("the folder holds %v", entries)
	}
	if !strings.Contains(read(t, dir), `"style": "digital"`) {
		t.Error("the file was not replaced")
	}
}

// FR-707: a folder that cannot be made is answered, not ignored.
func TestAFolderThatCannotBeMadeIsReported(t *testing.T) {
	t.Parallel()
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, nil, fileMode); err != nil {
		t.Fatal(err)
	}
	if err := New(filepath.Join(blocker, "TimeRibbon")).Save(settings.Defaults()); err == nil {
		t.Error("saving under a file succeeded")
	}
}

func TestAFaultReadingTheFileIsAnswered(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, FileName), folderMode); err != nil {
		t.Fatal(err)
	}
	if _, err := New(dir).Load(); err == nil {
		t.Error("a settings file that is a folder read without error")
	}
}

// compact answers text with its insignificant whitespace removed.
func compact(t *testing.T, text string) string {
	t.Helper()
	var out strings.Builder
	var value any
	if err := json.Unmarshal([]byte(text), &value); err != nil {
		t.Fatalf("not JSON: %v\n%s", err, text)
	}
	encoder := json.NewEncoder(&out)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(value)
	return strings.TrimSpace(out.String())
}
