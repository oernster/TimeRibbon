package application

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/oernster/timestrip/internal/domain/clock"
	"github.com/oernster/timestrip/internal/domain/settings"
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

// MoveClock moves the clock steps places, negative towards the start (FR-306).
func (s *Service) MoveClock(id string, steps int) error {
	return s.change(func(current settings.Settings) (settings.Settings, error) {
		return current.WithClockMoved(id, steps)
	})
}

// SearchPlaces answers the places whose default label, zone id or country contains query, ignoring
// case, ordered by label then zone (FR-302). An empty query answers every place.
func (s *Service) SearchPlaces(query string) []Place {
	wanted := strings.ToLower(strings.TrimSpace(query))
	var found []Place
	for _, place := range s.ports.Zones.Catalogue() {
		haystack := strings.ToLower(place.Label + "\n" + place.Zone + "\n" + place.Country)
		if strings.Contains(haystack, wanted) {
			found = append(found, place)
		}
	}
	slices.SortFunc(found, func(a, b Place) int {
		return cmp.Or(cmp.Compare(a.Label, b.Label), cmp.Compare(a.Zone, b.Zone))
	})
	return found
}

// known answers ErrUnknownZone, naming zone, when the tz database does not know it.
func (s *Service) known(zone string) error {
	if _, err := s.ports.Zones.Resolve(zone); err != nil {
		return fmt.Errorf("%w: %s", ErrUnknownZone, zone)
	}
	return nil
}
