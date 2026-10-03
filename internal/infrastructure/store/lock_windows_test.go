package store

import (
	"errors"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/oernster/timeribbon/internal/domain/settings"
)

// noSharing opens nothing to anyone else while the handle is held, as a backup tool, a sync client
// or a scanner may hold the file at sign-in.
const noSharing = 0

// holdUnshared opens path with no sharing and answers the function that lets it go.
func holdUnshared(t *testing.T, path string) func() {
	t.Helper()
	name, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		t.Fatal(err)
	}
	handle, err := syscall.CreateFile(name, syscall.GENERIC_READ, noSharing, nil, syscall.OPEN_EXISTING, syscall.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		t.Fatal(err)
	}
	return func() { _ = syscall.CloseHandle(handle) }
}

// As it happens at sign-in: the file is held open at launch, then let go before the first save. The clocks
// in it are never replaced by the defaults.
func TestAFileHeldOpenAtLaunchIsNeverSavedOver(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	write(t, dir, threeClocks)
	release := holdUnshared(t, filepath.Join(dir, FileName))
	store := New(dir)
	_, err := store.Load()
	release()
	if !errors.Is(err, ErrNotRead) {
		t.Errorf("load answered %v", err)
	}
	if err := store.Save(settings.Defaults()); !errors.Is(err, ErrNotRead) {
		t.Errorf("save answered %v", err)
	}
	if read(t, dir) != threeClocks {
		t.Error("the clocks were replaced by the defaults")
	}
}
