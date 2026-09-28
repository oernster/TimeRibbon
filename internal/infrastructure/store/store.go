// Package store keeps the settings in one indented JSON file a person can read (FR-701).
//
// Reading is tolerant: one bad value or one bad clock never costs the rest (FR-705). A file that is
// not JSON at all is kept aside under another name, never overwritten (FR-704). Writing replaces
// the file atomically (FR-702).
package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/oernster/timeribbon/internal/application"
	"github.com/oernster/timeribbon/internal/domain/settings"
)

// File names inside the settings folder.
const (
	FileName       = "settings.json"
	UnreadableName = "settings.unreadable.json"
	tempPattern    = "settings-*.tmp"
)

// formatVersion is written into every file so a later version can tell what it is reading.
const formatVersion = 1

// Permissions for what the store creates: the folder and the file are the user's alone.
const (
	folderMode fs.FileMode = 0o700
	fileMode   fs.FileMode = 0o600
)

// keptAsideNotice is shown when an unreadable file was renamed (FR-704).
const keptAsideNotice = "Settings could not be read; the old file was kept as " + UnreadableName

// ErrNotKeptAside is answered by Save when an unreadable file could not be renamed, so writing
// would destroy the only copy of the user's clocks.
var ErrNotKeptAside = errors.New("the unreadable settings file could not be kept aside, so it is not overwritten")

// Store is the application's Store port over one folder.
type Store struct {
	dir string
	// extras are top-level keys this version does not know, written back as found.
	extras []pair
	// blocked is set when an unreadable file could not be kept aside.
	blocked bool
}

// New answers a store over dir, normally %APPDATA%\TimeRibbon. Nothing is read or made until Load
// or Save is called.
func New(dir string) *Store {
	return &Store{dir: dir}
}

// Path answers the settings file's path.
func (s *Store) Path() string { return filepath.Join(s.dir, FileName) }

// Load reads the settings (FR-703 to FR-705). No file answers the defaults with no notice; a fault
// reading one that is there is answered as an error.
func (s *Store) Load() (application.Loaded, error) {
	raw, err := os.ReadFile(s.Path())
	if errors.Is(err, fs.ErrNotExist) {
		return application.Loaded{Settings: settings.Defaults()}, nil
	}
	if err != nil {
		return application.Loaded{Settings: settings.Defaults()}, fmt.Errorf("reading %s: %w", s.Path(), err)
	}
	decoded, extras, ok := decode(raw)
	if !ok {
		return s.keepAside()
	}
	s.extras = extras
	return application.Loaded{Settings: decoded}, nil
}

// keepAside renames an unreadable file and answers the defaults with the notice saying where it
// went. If the rename fails, saving is refused from then on.
func (s *Store) keepAside() (application.Loaded, error) {
	defaults := application.Loaded{Settings: settings.Defaults()}
	if err := os.Rename(s.Path(), filepath.Join(s.dir, UnreadableName)); err != nil {
		s.blocked = true
		return defaults, fmt.Errorf("%w: %w", ErrNotKeptAside, err)
	}
	defaults.Notice = keptAsideNotice
	return defaults, nil
}

// Save writes the settings to a temporary file in the same folder and renames it over the old
// one, so a failure part way leaves the previous file whole (FR-702).
func (s *Store) Save(current settings.Settings) error {
	if s.blocked {
		return ErrNotKeptAside
	}
	body, err := encode(current, s.extras)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(s.dir, folderMode); err != nil {
		return fmt.Errorf("making %s: %w", s.dir, err)
	}
	temp, err := os.CreateTemp(s.dir, tempPattern)
	if err != nil {
		return fmt.Errorf("writing in %s: %w", s.dir, err)
	}
	name := temp.Name()
	if err := writeAll(temp, body); err != nil {
		_ = os.Remove(name)
		return err
	}
	if err := os.Rename(name, s.Path()); err != nil {
		_ = os.Remove(name)
		return fmt.Errorf("replacing %s: %w", s.Path(), err)
	}
	return nil
}

// writeAll writes body, flushes it to the disk and closes the file, answering the first failure.
func writeAll(file *os.File, body []byte) error {
	_, writeErr := file.Write(body)
	syncErr := file.Sync()
	closeErr := file.Close()
	if err := errors.Join(writeErr, syncErr, closeErr); err != nil {
		return fmt.Errorf("writing %s: %w", file.Name(), err)
	}
	return os.Chmod(file.Name(), fileMode)
}

// pair is one top-level key with its value as written.
type pair struct {
	key   string
	value json.RawMessage
}
