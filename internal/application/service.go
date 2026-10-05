package application

import (
	"errors"
	"fmt"
	"sync"

	"github.com/oernster/timeribbon/internal/domain/settings"
	"github.com/oernster/timeribbon/ribbonkit/application/arranger"
	"github.com/oernster/timeribbon/ribbonkit/domain/placement"
)

// ErrUnknownZone is answered when a clock is set to a zone the tz database does not know.
var ErrUnknownZone = errors.New("unknown time zone")

// saveFailedPrefix begins the notice shown while the settings cannot be written (FR-707).
const saveFailedPrefix = "Settings could not be saved: "

// Layout is the size of one cell in each style and the padding around the cells, all in DIP. It
// has one home, the composition root; the front end draws cells at the sizes the snapshot hands it.
type Layout struct {
	Digital  placement.Size
	Analogue placement.Size
	// Prompt is the one cell an empty ribbon shows, holding the large Add clock button (FR-107).
	Prompt  placement.Size
	Padding int
	// HandleLane is the depth of the lane along the ribbon's inner side while the sun map is on: the
	// pull out's handle stands in it, so it covers no cell (FR-903).
	HandleLane int
}

// Layouts is the layout for each size (FR-610).
type Layouts struct {
	Large Layout
	Small Layout
}

// For answers the layout of size; the large one for a size it does not know.
func (l Layouts) For(size settings.Size) Layout {
	if size == settings.Small {
		return l.Small
	}
	return l.Large
}

// Service runs every use case over the current settings. It is safe to call from several
// goroutines: Wails, the tray and display events each call in on their own.
//
// The ribbon is arranged by the kit's Arranger, embedded so its use cases are the service's own;
// the service is its host, answering the clocks as the ribbon's content (see host).
type Service struct {
	*arranger.Arranger

	ports   Ports
	layouts Layouts

	mutex      sync.Mutex
	current    settings.Settings
	loadNotice string
	saveNotice string
	// measured is the cell width the page last measured its widest time and date to need, with the
	// choices it was measured for; the zero value, before it says, widens nothing (FR-620).
	measured Measured
}

// New answers a service over ports with the first-run settings; Start loads the stored ones.
func New(ports Ports, layouts Layouts) *Service {
	s := &Service{ports: ports, layouts: layouts, current: settings.Defaults()}
	s.Arranger = arranger.New(host{s}, ports.Monitors)
	return s
}

// Start loads the stored settings. A fault reading them is answered and the defaults are kept, so
// the ribbon still opens (FR-703, FR-704).
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
