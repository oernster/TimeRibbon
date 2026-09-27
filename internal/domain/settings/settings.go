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

// Size is how large every cell is drawn (FR-610).
type Size string

// The sizes.
const (
	Large Size = "large"
	Small Size = "small"
)

// Orientation is the direction cells run in (FR-103).
type Orientation string

// The orientations.
const (
	Horizontal Orientation = "horizontal"
	Vertical   Orientation = "vertical"
)

// homeEdges is each orientation's home edge (FR-409): a horizontal strip goes to the top, a
// vertical one to the right.
var homeEdges = map[Orientation]placement.Edge{Horizontal: placement.Top, Vertical: placement.Right}

// HomeEdge answers the edge a strip of orientation goes to when that orientation is chosen and
// wherever it has no place of its own (FR-403, FR-409); false for an orientation not offered.
func HomeEdge(orientation Orientation) (placement.Edge, bool) {
	edge, ok := homeEdges[orientation]
	return edge, ok
}

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
	Size        Size
	Format      clock.Format
	Orientation Orientation
	Theme       Theme
	AlwaysOnTop bool
	// Placement is where the strip was last left; nil until it has been placed (FR-403).
	Placement *placement.Stored
	// Clocks is the configured clocks in their order (FR-102).
	Clocks []Entry
}

// Defaults answers the settings of a first run (FR-703): digital, large (FR-610), 24-hour, vertical
// (FR-103, amended by Oliver on 2026-09-27), system theme, not on top, not yet placed, no clocks.
func Defaults() Settings {
	return Settings{
		Style:       Digital,
		Size:        Large,
		Format:      clock.TwentyFourHour,
		Orientation: Vertical,
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
	if s.Size != Large && s.Size != Small {
		s.Size = defaults.Size
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
