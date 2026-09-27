// Package startup reads and writes the Start with Windows value under HKCU (FR-605, FR-805).
// Settings and setup both go through it, so the two write the same single value.
package startup

import (
	"errors"
	"fmt"

	"golang.org/x/sys/windows/registry"

	"github.com/oernster/timestrip/internal/product"
)

// RunKey is the per-user key Windows starts programs from at sign-in.
const RunKey = `Software\Microsoft\Windows\CurrentVersion\Run`

// ValueName is the value TimeStrip writes under RunKey.
const ValueName = product.Name

// Entry is the application's StartupEntry port over one value.
type Entry struct {
	key     string
	value   string
	program string
}

// New answers the entry for program, the full path of TimeStrip's executable, under RunKey.
func New(program string) Entry {
	return Entry{key: RunKey, value: ValueName, program: program}
}

// Command answers what the value holds: the quoted program path and no arguments, so a sign-in
// start shows the strip as a normal launch does (FR-605).
func (e Entry) Command() string {
	return `"` + e.program + `"`
}

// Enabled answers whether the value is present.
func (e Entry) Enabled() (bool, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, e.key, registry.QUERY_VALUE)
	if errors.Is(err, registry.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("opening %s: %w", e.key, err)
	}
	defer key.Close()
	_, _, err = key.GetStringValue(e.value)
	if errors.Is(err, registry.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("reading %s: %w", e.value, err)
	}
	return true, nil
}

// Enable writes the value.
func (e Entry) Enable() error {
	key, _, err := registry.CreateKey(registry.CURRENT_USER, e.key, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("opening %s: %w", e.key, err)
	}
	defer key.Close()
	if err := key.SetStringValue(e.value, e.Command()); err != nil {
		return fmt.Errorf("writing %s: %w", e.value, err)
	}
	return nil
}

// Disable deletes the value; one that is already gone is not an error.
func (e Entry) Disable() error {
	key, err := registry.OpenKey(registry.CURRENT_USER, e.key, registry.SET_VALUE)
	if errors.Is(err, registry.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("opening %s: %w", e.key, err)
	}
	defer key.Close()
	if err := key.DeleteValue(e.value); err != nil && !errors.Is(err, registry.ErrNotExist) {
		return fmt.Errorf("deleting %s: %w", e.value, err)
	}
	return nil
}
