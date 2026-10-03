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
	"strconv"

	"github.com/oernster/timeribbon/internal/application"
	"github.com/oernster/timeribbon/internal/domain/settings"
)

// File names inside the settings folder.
const (
	FileName           = "settings.json"
	keptAsideStem      = "settings.unreadable"
	keptAsideExtension = ".json"
	UnreadableName     = keptAsideStem + keptAsideExtension
	tempPattern        = "settings-*.tmp"
)

// keptAsideLimit bounds how many damaged files are kept aside. Once every name is taken the next is
// not renamed and saving is refused, so even then nothing is overwritten.
const keptAsideLimit = 100

// formatVersion is written into every file so a later version can tell what it is reading.
const formatVersion = 1

// Permissions for what the store creates: the folder and the file are the user's alone.
const (
	folderMode fs.FileMode = 0o700
	fileMode   fs.FileMode = 0o600
)

// keptAsideNotice is shown when an unreadable file was renamed to name (FR-704).
func keptAsideNotice(name string) string {
	return "Settings could not be read; the old file was kept as " + name
}

// keptAsideName answers the name the nth kept-aside file takes: UnreadableName for the first, then
// the same name numbered from 2, so an earlier copy is never the one replaced (FR-704).
func keptAsideName(n int) string {
	if n <= 1 {
		return UnreadableName
	}
	return keptAsideStem + "-" + strconv.Itoa(n) + keptAsideExtension
}

// ErrNotRead is answered by Load and then by every Save when the file was there but could not be
// read: held open by another program, say. The defaults standing in for it are not the user's, so
// nothing is saved over it until the next run reads it (FR-704).
var ErrNotRead = errors.New("nothing is saved over it until TimeRibbon is started again and reads it")

// ErrNotKeptAside is answered by Save when an unreadable file could not be renamed, so writing
// would destroy the only copy of the user's clocks.
var ErrNotKeptAside = errors.New("the unreadable settings file could not be kept aside, so it is not overwritten")

// Store is the application's Store port over one folder.
type Store struct {
	dir string
	// extras are top-level keys this version does not know, written back as found.
	extras []pair
	// blocked is why saving is refused: ErrNotRead or ErrNotKeptAside beneath; nil while it is not.
	blocked error
}

// New answers a store over dir, normally %APPDATA%\TimeRibbon. Nothing is read or made until Load
// or Save is called.
func New(dir string) *Store {
	return &Store{dir: dir}
}

// Path answers the settings file's path.
func (s *Store) Path() string { return filepath.Join(s.dir, FileName) }

// Load reads the settings (FR-703 to FR-705). No file answers the defaults with no notice. A fault
// reading one that is there is answered as an error and refuses every later save (FR-704).
func (s *Store) Load() (application.Loaded, error) {
	raw, err := os.ReadFile(s.Path())
	if errors.Is(err, fs.ErrNotExist) {
		return application.Loaded{Settings: settings.Defaults()}, nil
	}
	if err != nil {
		s.blocked = fmt.Errorf("reading %s: %w; %w", s.Path(), err, ErrNotRead)
		return application.Loaded{Settings: settings.Defaults()}, s.blocked
	}
	decoded, extras, ok := decode(raw)
	if !ok {
		return s.keepAside()
	}
	s.extras = extras
	return application.Loaded{Settings: decoded}, nil
}

// keepAside renames an unreadable file to the first kept-aside name not yet taken and answers the
// defaults with the notice saying where it went; an earlier copy is never replaced (FR-704). If no
// name is free or the rename fails, saving is refused from then on.
func (s *Store) keepAside() (application.Loaded, error) {
	defaults := application.Loaded{Settings: settings.Defaults()}
	name, err := s.freeKeptAsideName()
	if err == nil {
		err = os.Rename(s.Path(), filepath.Join(s.dir, name))
	}
	if err != nil {
		s.blocked = fmt.Errorf("%w: %w", ErrNotKeptAside, err)
		return defaults, s.blocked
	}
	defaults.Notice = keptAsideNotice(name)
	return defaults, nil
}

// errNoFreeName is answered when every kept-aside name is taken.
var errNoFreeName = errors.New("every name a damaged file is kept under is taken")

// freeKeptAsideName answers the first kept-aside name nothing in the folder holds.
func (s *Store) freeKeptAsideName() (string, error) {
	for n := 1; n <= keptAsideLimit; n++ {
		name := keptAsideName(n)
		_, err := os.Lstat(filepath.Join(s.dir, name))
		if errors.Is(err, fs.ErrNotExist) {
			return name, nil
		}
		if err != nil {
			return "", err
		}
	}
	return "", errNoFreeName
}

// Save writes the settings to a temporary file in the same folder and renames it over the old
// one, so a failure part way leaves the previous file whole (FR-702).
func (s *Store) Save(current settings.Settings) error {
	if s.blocked != nil {
		return s.blocked
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
