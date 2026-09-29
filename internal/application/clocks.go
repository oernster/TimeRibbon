package application

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/oernster/timeribbon/internal/domain/clock"
	"github.com/oernster/timeribbon/internal/domain/settings"
)

// AddClock appends a clock for zone under its default label and answers the new clock's id
// (FR-301). A zone the tz database does not know is refused and nothing changes.
func (s *Service) AddClock(zone string) (string, error) {
	if err := s.known(zone); err != nil {
		return "", err
	}
	id := s.ports.IDs.NewID()
	entry := settings.Entry{ID: id, Zone: zone, Label: clock.DefaultLabel(zone)}
	err := s.change(func(current settings.Settings) (settings.Settings, error) {
		return current.WithClockAdded(entry), nil
	})
	return id, err
}

// RenameClock stores typed as the clock's label, trimmed and capped; the default label when
// nothing is left (FR-303, FR-307).
func (s *Service) RenameClock(id, typed string) error {
	return s.change(func(current settings.Settings) (settings.Settings, error) {
		entry, err := current.Clock(id)
		if err != nil {
			return current, err
		}
		entry.Label = clock.Label(typed, entry.Zone)
		return current.WithClockReplaced(entry)
	})
}

// RezoneClock points the clock at zone, keeping its place and any label the user typed (FR-304).
// It is also how an invalid clock is repaired (FR-706).
func (s *Service) RezoneClock(id, zone string) error {
	if err := s.known(zone); err != nil {
		return err
	}
	return s.change(func(current settings.Settings) (settings.Settings, error) {
		entry, err := current.Clock(id)
		if err != nil {
			return current, err
		}
		return current.WithClockReplaced(settings.Rezoned(entry, zone))
	})
}

// RemoveClock removes the clock and closes the gap (FR-305). The confirmation is the window's.
func (s *Service) RemoveClock(id string) error {
	return s.change(func(current settings.Settings) (settings.Settings, error) {
		return current.WithoutClock(id)
	})
}

// How well a place matches a query, best first (FR-302, Amendment 29): its label begins with the
// query; a later word of its label does; a word of its country or zone does. A query found only
// inside a word is no match, so "l" lists London and not Adelaide.
const (
	labelBegins = iota
	labelWordBegins
	elsewhereWordBegins
	noMatch
)

// SearchPlaces answers the places a word of whose default label, country or zone id begins with
// query, ignoring case: the best matches first, then by label then zone (FR-302). An empty query
// answers every place.
func (s *Service) SearchPlaces(query string) []Place {
	wanted := strings.ToLower(strings.TrimSpace(query))
	type ranked struct {
		place Place
		rank  int
	}
	var found []ranked
	for _, place := range s.ports.Zones.Catalogue() {
		if rank := matchRank(place, wanted); rank != noMatch {
			found = append(found, ranked{place, rank})
		}
	}
	slices.SortFunc(found, func(a, b ranked) int {
		return cmp.Or(cmp.Compare(a.rank, b.rank), cmp.Compare(a.place.Label, b.place.Label), cmp.Compare(a.place.Zone, b.place.Zone))
	})
	places := make([]Place, 0, len(found))
	for _, each := range found {
		places = append(places, each.place)
	}
	return places
}

// matchRank answers how well place matches wanted, which is already lower case.
func matchRank(place Place, wanted string) int {
	label := strings.ToLower(place.Label)
	switch {
	case strings.HasPrefix(label, wanted):
		return labelBegins
	case beginsAWord(label, wanted):
		return labelWordBegins
	case beginsAWord(strings.ToLower(place.Country), wanted), beginsAWord(strings.ToLower(place.Zone), wanted):
		return elsewhereWordBegins
	}
	return noMatch
}

// beginsAWord answers whether wanted appears in text where a word begins: at its start or just
// after a character that is neither a letter nor a digit, such as a space, "/", "_" or "(".
func beginsAWord(text, wanted string) bool {
	for from := 0; ; {
		at := strings.Index(text[from:], wanted)
		if at < 0 {
			return false
		}
		at += from
		before, _ := utf8.DecodeLastRuneInString(text[:at])
		if at == 0 || !(unicode.IsLetter(before) || unicode.IsDigit(before)) {
			return true
		}
		from = at + 1
	}
}

// known answers ErrUnknownZone, naming zone, when the tz database does not know it.
func (s *Service) known(zone string) error {
	if _, err := s.ports.Zones.Resolve(zone); err != nil {
		return fmt.Errorf("%w: %s", ErrUnknownZone, zone)
	}
	return nil
}
