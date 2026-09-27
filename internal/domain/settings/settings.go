// Package settings holds the user's choices as one value and the operations that change them.
//
// Every operation answers a new value and leaves the receiver as it was, so a change that is
// refused changes nothing. Derived values (offsets, abbreviations, times) are never held here
// (FR-701).
package settings

import (
	"errors"
	"slices"

	"github.com/oernster/timestrip/internal/domain/clock"
	"github.com/oernster/timestrip/internal/domain/placement"
)

// Style is how every cell presents its time (FR-603, FR-604).
type Style string

// The styles. The string values are what the settings file holds.
const (
	Digital  Style = "digital"
	Analogue Style = "analogue"
)

// Orientation is the direction cells run in (FR-103).
type Orientation string

// The orientations.
const (
	Horizontal Orientation = "horizontal"
	Vertical   Orientation = "vertical"
)

// Theme is the colour scheme (FR-606).
type Theme string

// The themes.
const (
	System Theme = "system"
	Light  Theme = "light"
	Dark   Theme = "dark"
)

// ErrNoSuchClock is answered when an operation names a clock id that is not configured.
var ErrNoSuchClock = errors.New("no clock has that id")

// Entry is one configured clock as stored. Its position is its index in Settings.Clocks.
type Entry struct {
	// ID is stable for the clock's lifetime and never reused.
	ID string
	// Zone is the IANA zone id as stored; it may name a zone the tz database does not know.
	Zone string
	// Label is the place name shown.
	Label string
	// Unreadable is the reason the stored entry could not be read; empty when it was read.
	Unreadable string
	// Original is the entry's stored text, kept so an unreadable entry is written back as it was
	// found (FR-705).
	Original string
}

// Settings is every choice the user has made.
type Settings struct {
	Style       Style
	Format      clock.Format
	Orientation Orientation
	Theme       Theme
	AlwaysOnTop bool
	// Placement is where the strip was last left; nil until it has been placed (FR-403).
	Placement *placement.Stored
	// Clocks is the configured clocks in their order (FR-102).
	Clocks []Entry
}

// Defaults answers the settings of a first run (FR-703): digital, 24-hour, horizontal, system
// theme, not on top, not yet placed, no clocks.
func Defaults() Settings {
	return Settings{
		Style:       Digital,
		Format:      clock.TwentyFourHour,
		Orientation: Horizontal,
		Theme:       System,
	}
}

// Normalised answers the settings with any choice that is not one of the known values replaced by
// its default, so a hand-edited file holding a word it should not cannot leave a choice unset.
func (s Settings) Normalised() Settings {
	defaults := Defaults()
	if s.Style != Digital && s.Style != Analogue {
		s.Style = defaults.Style
	}
	if s.Format != clock.TwentyFourHour && s.Format != clock.TwelveHour {
		s.Format = defaults.Format
	}
	if s.Orientation != Horizontal && s.Orientation != Vertical {
		s.Orientation = defaults.Orientation
	}
	if s.Theme != System && s.Theme != Light && s.Theme != Dark {
		s.Theme = defaults.Theme
	}
	s.Clocks = slices.Clone(s.Clocks)
	return s
}

// WithClockAdded answers the settings with entry appended at the end (FR-301).
func (s Settings) WithClockAdded(entry Entry) Settings {
	s.Clocks = append(slices.Clone(s.Clocks), entry)
	return s
}

// WithoutClock answers the settings with the clock id removed and the gap closed (FR-305).
func (s Settings) WithoutClock(id string) (Settings, error) {
	index := s.indexOf(id)
	if index < 0 {
		return s, ErrNoSuchClock
	}
	s.Clocks = slices.Delete(slices.Clone(s.Clocks), index, index+1)
	return s, nil
}

// WithClockMoved answers the settings with the clock id moved by steps places, negative towards
// the start, stopping at either end (FR-306).
func (s Settings) WithClockMoved(id string, steps int) (Settings, error) {
	from := s.indexOf(id)
	if from < 0 {
		return s, ErrNoSuchClock
	}
	to := min(max(from+steps, 0), len(s.Clocks)-1)
	moved := slices.Delete(slices.Clone(s.Clocks), from, from+1)
	s.Clocks = slices.Insert(moved, to, s.Clocks[from])
	return s, nil
}

// WithClockReplaced answers the settings with the clock of entry's id replaced by entry, keeping
// its place in the order (FR-303, FR-304).
func (s Settings) WithClockReplaced(entry Entry) (Settings, error) {
	index := s.indexOf(entry.ID)
	if index < 0 {
		return s, ErrNoSuchClock
	}
	s.Clocks = slices.Clone(s.Clocks)
	s.Clocks[index] = entry
	return s, nil
}

// Clock answers the configured clock with id.
func (s Settings) Clock(id string) (Entry, error) {
	index := s.indexOf(id)
	if index < 0 {
		return Entry{}, ErrNoSuchClock
	}
	return s.Clocks[index], nil
}

func (s Settings) indexOf(id string) int {
	return slices.IndexFunc(s.Clocks, func(entry Entry) bool { return entry.ID == id })
}

// Rezoned answers entry pointed at zone (FR-304). Its label follows the zone only where it was the
// old zone's default label; a label the user typed is kept. An entry that was unreadable becomes
// readable, since its zone and label are now the user's own choice.
func Rezoned(entry Entry, zone string) Entry {
	if entry.Label == clock.DefaultLabel(entry.Zone) || entry.Label == "" {
		entry.Label = clock.DefaultLabel(zone)
	}
	entry.Zone = zone
	entry.Unreadable = ""
	entry.Original = ""
	return entry
}
