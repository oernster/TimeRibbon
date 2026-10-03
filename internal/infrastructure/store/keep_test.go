package store

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/oernster/timeribbon/internal/domain/settings"
)

// threeClocks is a file holding three clocks, the user's own.
const threeClocks = `{"clocks": [
	{"id": "m", "zone": "Australia/Sydney", "label": "Mum"},
	{"id": "k", "zone": "Asia/Tokyo", "label": "Kenji"},
	{"id": "o", "zone": "Europe/London", "label": "Office"}
]}`

// A file that is there but cannot be read is never saved over, even once it can be read again: the
// defaults standing in for it are not the user's (FR-704).
func TestAFileThatCouldNotBeReadIsNeverSavedOver(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	blocker := filepath.Join(dir, FileName)
	if err := os.Mkdir(blocker, folderMode); err != nil {
		t.Fatal(err)
	}
	store := New(dir)
	if _, err := store.Load(); !errors.Is(err, ErrNotRead) {
		t.Errorf("load answered %v", err)
	}
	if err := os.Remove(blocker); err != nil {
		t.Fatal(err)
	}
	write(t, dir, threeClocks)
	if err := store.Save(settings.Defaults()); !errors.Is(err, ErrNotRead) {
		t.Errorf("save answered %v", err)
	}
	if read(t, dir) != threeClocks {
		t.Error("the file that could not be read was saved over")
	}
}

// A second damaged file never replaces the copy kept aside from the first (FR-704).
func TestASecondDamagedFileKeepsTheFirstCopy(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	first := strings.TrimSuffix(threeClocks, "]}")
	write(t, dir, first)
	if loaded, err := New(dir).Load(); err != nil || loaded.Notice != keptAsideNotice(UnreadableName) {
		t.Fatalf("first: %+v (%v)", loaded, err)
	}
	second := `{"clocks": [ PARTIAL`
	write(t, dir, second)
	loaded, err := New(dir).Load()
	secondName := keptAsideName(2)
	if err != nil || loaded.Notice != keptAsideNotice(secondName) {
		t.Errorf("second: %+v (%v)", loaded, err)
	}
	for name, want := range map[string]string{UnreadableName: first, secondName: second} {
		kept, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil || string(kept) != want {
			t.Errorf("%s holds %q (%v), want %q", name, kept, err, want)
		}
	}
}

// Two clocks sharing an id are both kept, each answering to an id of its own, so a change to one
// never reaches the other (FR-705).
func TestClocksSharingAnIdAreToldApart(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	write(t, dir, `{"clocks": [
		{"id": "same", "zone": "Europe/London", "label": "London"},
		{"id": "same", "zone": "Asia/Tokyo", "label": "Tokyo"}
	]}`)
	store := New(dir)
	loaded, err := store.Load()
	if err != nil || len(loaded.Settings.Clocks) != 2 {
		t.Fatalf("loaded %+v (%v)", loaded, err)
	}
	london, tokyo := loaded.Settings.Clocks[0], loaded.Settings.Clocks[1]
	if london.ID != "same" || tokyo.ID == london.ID {
		t.Fatalf("ids %q and %q", london.ID, tokyo.ID)
	}
	tokyo.Label = "Kenji at work"
	renamed, err := loaded.Settings.WithClockReplaced(tokyo)
	if err != nil || renamed.Clocks[0].Label != "London" || renamed.Clocks[1].Label != "Kenji at work" {
		t.Errorf("renaming Tokyo left %+v (%v)", renamed.Clocks, err)
	}
	removed, err := loaded.Settings.WithoutClock(tokyo.ID)
	if err != nil || len(removed.Clocks) != 1 || removed.Clocks[0].Label != "London" {
		t.Errorf("removing Tokyo left %+v (%v)", removed.Clocks, err)
	}
	if err := store.Save(loaded.Settings); err != nil {
		t.Fatal(err)
	}
	if again, _ := New(dir).Load(); again.Settings.Clocks[1].ID != tokyo.ID {
		t.Errorf("Tokyo's own id was not saved: %+v", again.Settings.Clocks)
	}
}

// An unreadable entry's id for the session never collides with an id the file holds (FR-705).
func TestAnUnreadableEntryNeverTakesAWrittenId(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	written := unreadableIDPrefix + "1"
	write(t, dir, `{"clocks": [{"id": "`+written+`", "zone": "Europe/London"}, {"id": 7}]}`)
	loaded, err := New(dir).Load()
	if err != nil || len(loaded.Settings.Clocks) != 2 {
		t.Fatalf("loaded %+v (%v)", loaded, err)
	}
	var ids []string
	for _, entry := range loaded.Settings.Clocks {
		ids = append(ids, entry.ID)
	}
	if ids[0] != written || slices.Contains(ids[1:], written) {
		t.Errorf("ids %v", ids)
	}
}

// A file saved with a UTF-8 byte order mark, as Notepad and PowerShell 5 can, reads whole (FR-701).
func TestAByteOrderMarkIsRead(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	write(t, dir, byteOrderMark+`{"style": "analogue", "later": 1, "clocks": [{"id": "a", "zone": "Europe/London"}]}`)
	store := New(dir)
	loaded, err := store.Load()
	if err != nil || loaded.Notice != "" || loaded.Settings.Style != settings.Analogue || len(loaded.Settings.Clocks) != 1 {
		t.Fatalf("loaded %+v (%v)", loaded, err)
	}
	if err := store.Save(loaded.Settings); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(read(t, dir), `"later": 1`) {
		t.Errorf("an unknown key was lost:\n%s", read(t, dir))
	}
}
