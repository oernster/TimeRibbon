package startup

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/oernster/timeribbon/internal/product"
)

// Ported from o7Debrief's LinuxAutostart, where both traps below were met on a real Ubuntu desktop.
//
// Inside a Flatpak, XDG_CONFIG_HOME names the app's private configuration, which the session never
// reads: an entry written there is created without error and read back as on, while nothing starts.
// So under a Flatpak the entry goes to the real ~/.config/autostart, which the manifest's
// xdg-config/autostart:create grant mounts at its true path. The session runs the entry's command
// from outside the sandbox, so for a Flatpak that command is flatpak run with the app's id.

const (
	configHomeVariable = "XDG_CONFIG_HOME"
	flatpakIDVariable  = "FLATPAK_ID"
	configFolder       = ".config"
	autostartFolder    = "autostart"
	entrySuffix        = ".desktop"
	folderMode         = fs.FileMode(0o700)
	entryMode          = fs.FileMode(0o644)
)

// The lines that mark an entry switched off. The desktop honours either over the file merely
// existing, so an entry carrying one is reported off rather than letting Settings claim otherwise.
const (
	hiddenLine        = "Hidden=true"
	gnomeDisabledLine = "X-GNOME-Autostart-enabled=false"
)

// entryTemplate is the entry written. X-GNOME-Autostart-enabled is stated so an older entry that
// switched it off cannot outlive being switched back on.
const entryTemplate = `[Desktop Entry]
Type=Application
Name=%s
Exec=%s
Icon=%s
Terminal=false
NoDisplay=true
X-GNOME-Autostart-enabled=true
`

// errNoHome is answered when the session names no home folder, so there is nowhere to write.
var errNoHome = errors.New("the session names no home folder, so there is no autostart folder")

// Entry is the application's StartupEntry port over one XDG autostart entry.
type Entry struct {
	dir     string
	command string
	problem error
}

// New answers the entry for program, the full path of TimeRibbon's executable, in the session's
// autostart folder. Under a Flatpak the command is flatpak run with its id instead.
func New(program string) Entry {
	dir, err := autostartDir(os.Getenv, os.UserHomeDir)
	entry := At(dir, program)
	if id := os.Getenv(flatpakIDVariable); id != "" {
		entry.command = "flatpak run " + id
	}
	entry.problem = err
	return entry
}

// At answers the entry for program in dir. Only a test names a folder of its own, so the real
// session's entries are never touched.
func At(dir, program string) Entry {
	return Entry{dir: dir, command: execQuoted(program)}
}

// Command answers what the entry runs: the program with no arguments, so a sign-in start shows the
// ribbon as a normal launch does (FR-605).
func (e Entry) Command() string { return e.command }

// path is the entry's file, named by the app id as the desktop convention has it.
func (e Entry) path() string { return filepath.Join(e.dir, product.AppID+entrySuffix) }

// Enabled answers whether the entry is present and not switched off. A missing entry is off; one
// that is there and cannot be read is a fault, answered as one.
func (e Entry) Enabled() (bool, error) {
	if e.problem != nil {
		return false, e.problem
	}
	raw, err := os.ReadFile(e.path())
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("reading %s: %w", e.path(), err)
	}
	text := string(raw)
	return !strings.Contains(text, hiddenLine) && !strings.Contains(text, gnomeDisabledLine), nil
}

// Enable writes the entry, making the autostart folder where it is missing.
func (e Entry) Enable() error {
	if e.problem != nil {
		return e.problem
	}
	if err := os.MkdirAll(e.dir, folderMode); err != nil {
		return fmt.Errorf("making %s: %w", e.dir, err)
	}
	text := fmt.Sprintf(entryTemplate, product.Name, e.command, product.AppID)
	if err := os.WriteFile(e.path(), []byte(text), entryMode); err != nil {
		return fmt.Errorf("writing %s: %w", e.path(), err)
	}
	return nil
}

// Disable removes the entry; one that is already gone is not an error. Removal rather than a
// Hidden line, so nothing is left behind for a later version to misread.
func (e Entry) Disable() error {
	if e.problem != nil {
		return e.problem
	}
	if err := os.Remove(e.path()); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("removing %s: %w", e.path(), err)
	}
	return nil
}

// autostartDir answers the folder the session starts programs from. Under a Flatpak it ignores
// XDG_CONFIG_HOME, which names the sandbox's private configuration there.
func autostartDir(getenv func(string) string, home func() (string, error)) (string, error) {
	if base := getenv(configHomeVariable); base != "" && getenv(flatpakIDVariable) == "" {
		return filepath.Join(base, autostartFolder), nil
	}
	dir, err := home()
	if err != nil || dir == "" {
		return "", errNoHome
	}
	return filepath.Join(dir, configFolder, autostartFolder), nil
}

// execQuoted quotes a path for an Exec line. Inside the quotes the characters the Desktop Entry
// specification reserves are escaped with a backslash; then the value as a whole is escaped as a
// string value, which doubles every backslash.
func execQuoted(path string) string {
	quoted := `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`, "`", "\\`", `$`, `\$`).Replace(path) + `"`
	return strings.ReplaceAll(quoted, `\`, `\\`)
}
