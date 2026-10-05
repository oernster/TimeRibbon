// Package settings holds the user's choices as one value and the operations that change them.
//
// Every operation answers a new value and leaves the receiver as it was, so a change that is
// refused changes nothing. Derived values (offsets, abbreviations, times) are never held here
// (FR-701).
package settings

import (
	"errors"
	"slices"

	"github.com/oernster/ribbonkit/domain/localtime"
	"github.com/oernster/ribbonkit/domain/ribbon"
	"github.com/oernster/timeribbon/internal/domain/clock"
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

// Settings is every choice the user has made: the ribbon's own, which ribbonkit holds, then the
// clocks'. The ribbon's are embedded, so they read as fields of Settings.
type Settings struct {
	ribbon.Choices
	Style      Style
	Size       Size
	Format     localtime.Format
	DateFormat clock.DateFormat
	// SunMap shows the world map lit by day beside the ribbon (FR-901); PullOut keeps a vertical
	// ribbon's map pulled out (FR-903). Both are off on a first run.
	SunMap  bool
	PullOut bool
	// Clocks is the configured clocks in their order (FR-102).
	Clocks []Entry
}

// Defaults answers the settings of a first run (FR-703): the ribbon's own (vertical, FR-103, amended
// by Oliver on 2026-09-27; pinned, FR-613), then digital, large (FR-610), 24-hour, no clocks.
func Defaults() Settings {
	return Settings{
		Choices:    ribbon.Defaults(),
		Style:      Digital,
		Size:       Large,
		Format:     localtime.TwentyFourHour,
		DateFormat: clock.DayMonth,
	}
}

// Normalised answers the settings with any choice that is not one of the known values replaced by
// its default, so a hand-edited file holding a word it should not cannot leave a choice unset.
func (s Settings) Normalised() Settings {
	defaults := Defaults()
	s.Choices = s.Choices.Normalised()
	if s.Style != Digital && s.Style != Analogue {
		s.Style = defaults.Style
	}
	if s.Size != Large && s.Size != Small {
		s.Size = defaults.Size
	}
	if !slices.Contains(localtime.Formats, s.Format) {
		s.Format = defaults.Format
	}
	if !slices.Contains(clock.DateFormats, s.DateFormat) {
		s.DateFormat = defaults.DateFormat
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
