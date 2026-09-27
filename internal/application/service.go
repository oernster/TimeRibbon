package application

import (
	"errors"
	"fmt"
	"sync"

	"github.com/oernster/timestrip/internal/domain/placement"
	"github.com/oernster/timestrip/internal/domain/settings"
)

// ErrUnknownZone is answered when a clock is set to a zone the tz database does not know.
var ErrUnknownZone = errors.New("unknown time zone")

// ErrUnknownChoice is answered when a setting is given a value it does not offer.
var ErrUnknownChoice = errors.New("not one of the values this setting offers")

// ErrNoMonitors is answered when Windows reports no display at all.
var ErrNoMonitors = errors.New("no display is reported")

// ErrNegativeLength is answered when a length that cannot be negative is given as one.
var ErrNegativeLength = errors.New("a length cannot be negative")

// saveFailedPrefix begins the notice shown while the settings cannot be written (FR-707).
const saveFailedPrefix = "Settings could not be saved: "

// Layout is the size of one cell in each style and the padding around the cells, all in DIP. It
// has one home, the composition root; the front end draws cells at the sizes the snapshot hands it.
type Layout struct {
	Digital  placement.Size
	Analogue placement.Size
	// Prompt is the one cell an empty strip shows, holding the large Add clock button (FR-107).
	Prompt  placement.Size
	Padding int
}

// Service runs every use case over the current settings. It is safe to call from several
// goroutines: Wails, the tray and display events each call in on their own.
type Service struct {
	ports  Ports
	layout Layout

	mutex      sync.Mutex
	current    settings.Settings
	loadNotice string
	saveNotice string
	// scrollbar is the thickness in DIP of the scroll bar the page draws, as the page measured it;
	// zero until it says (FR-106).
	scrollbar int
	// arranged is the strip's length when it was last arranged, so a change of length can be told
	// from anything else that arranges it (FR-104).
	arranged stripLength
}

// stripLength is the strip's length in DIP along its orientation; known is false until the strip
// has been arranged once, when there is nothing yet for a length to differ from.
type stripLength struct {
	known    bool
	vertical bool
	length   int
}

// SetScrollbar records the thickness in DIP of the scroll bar the page draws, which a scrolling
// strip makes room for across its cells (FR-106). Only the page can measure it: it is the web
// engine's bar, not one Windows reports.
func (s *Service) SetScrollbar(dip int) error {
	if dip < 0 {
		return fmt.Errorf("%w: a scroll bar of %d", ErrNegativeLength, dip)
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.scrollbar = dip
	return nil
}

// New answers a service over ports with the first-run settings; Start loads the stored ones.
func New(ports Ports, layout Layout) *Service {
	return &Service{ports: ports, layout: layout, current: settings.Defaults()}
}

// Start loads the stored settings. A fault reading them is answered and the defaults are kept, so
// the strip still opens (FR-703, FR-704).
func (s *Service) Start() error {
	loaded, err := s.ports.Store.Load()
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if err != nil {
		s.loadNotice = fmt.Sprintf("Settings could not be read: %v", err)
		return err
	}
	s.current = loaded.Settings.Normalised()
	s.loadNotice = loaded.Notice
	return nil
}

// Settings answers a copy of the current settings.
func (s *Service) Settings() settings.Settings {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.current.Normalised()
}

// DismissNotices clears the notices the user has read. A save failure that still holds is raised
// again by the next save that fails.
func (s *Service) DismissNotices() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.loadNotice = ""
	s.saveNotice = ""
}

// change applies edit to the current settings and saves the result. An edit that answers an error
// changes nothing. A save that fails keeps the change in effect and raises a notice until a later
// save succeeds (FR-602, FR-707); it is answered too, so a caller can tell.
func (s *Service) change(edit func(settings.Settings) (settings.Settings, error)) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	next, err := edit(s.current)
	if err != nil {
		return err
	}
	s.current = next
	if err := s.ports.Store.Save(next); err != nil {
		s.saveNotice = saveFailedPrefix + err.Error()
		return err
	}
	s.saveNotice = ""
	return nil
}

// notices answers the notices standing now, the load notice first. The caller holds the mutex.
func (s *Service) notices() []string {
	var out []string
	for _, notice := range []string{s.loadNotice, s.saveNotice} {
		if notice != "" {
			out = append(out, notice)
		}
	}
	return out
}
