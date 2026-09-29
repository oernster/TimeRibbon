package application

import (
	"errors"
	"fmt"
	"math"
	"sync"

	"github.com/oernster/timeribbon/internal/domain/placement"
	"github.com/oernster/timeribbon/internal/domain/settings"
)

// ErrUnknownZone is answered when a clock is set to a zone the tz database does not know.
var ErrUnknownZone = errors.New("unknown time zone")

// ErrUnknownChoice is answered when a setting is given a value it does not offer.
var ErrUnknownChoice = errors.New("not one of the values this setting offers")

// ErrNoMonitors is answered when Windows reports no display at all.
var ErrNoMonitors = errors.New("no display is reported")

// ErrNotAgainstAnEdge is answered when the tab of a ribbon flush against no edge is asked for; such a
// ribbon never collapses (FR-619).
var ErrNotAgainstAnEdge = errors.New("the ribbon stands against no edge, so it has no tab")

// ErrNegativeLength is answered when a length that cannot be negative is given as one.
var ErrNegativeLength = errors.New("a length cannot be negative")

// ErrUnusableScale is answered when the page reports a scale that is not a positive number.
var ErrUnusableScale = errors.New("a scale must be a positive number")

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
type Service struct {
	ports   Ports
	layouts Layouts

	mutex      sync.Mutex
	current    settings.Settings
	loadNotice string
	saveNotice string
	// scrollbar is the thickness in DIP of the scroll bar the page draws, as the page measured it;
	// zero until it says (FR-106).
	scrollbar int
	// pixelsPerDIP is the scale the page is really drawn at, in window pixels to each DIP, as the
	// page reported it; zero until it says, when the display's DPI stands in for it.
	pixelsPerDIP float64
	// arranged is the ribbon's length when it was last arranged, so a change of length can be told
	// from anything else that arranges it (FR-104).
	arranged ribbonLength
	// last is where the ribbon was last arranged, so a ribbon placed again can keep the edge it lay
	// against (FR-408, FR-610).
	last lastPlaced
}

// lastPlaced is where the ribbon was last arranged: the display, its corner and its size, in that
// display's pixels; known is false until it has been arranged once.
type lastPlaced struct {
	known  bool
	device string
	at     placement.Point
	size   placement.Size
}

// ribbonLength is the ribbon's length in DIP along its orientation; known is false until the ribbon
// has been arranged once, when there is nothing yet for a length to differ from.
type ribbonLength struct {
	known    bool
	vertical bool
	length   int
}

// SetScrollbar records the thickness in DIP of the scroll bar the page draws, which a scrolling
// ribbon makes room for across its cells (FR-106). Only the page can measure it: it is the web
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

// SetPixelsPerDIP records the scale the page is really drawn at, in window pixels to each DIP.
// Every size is converted with it from then on rather than with the display's DPI, which Windows'
// text size leaves as it was while it enlarges the page, so the window always fits the page.
func (s *Service) SetPixelsPerDIP(scale float64) error {
	if !(scale > 0) || math.IsInf(scale, 1) {
		return fmt.Errorf("%w: %v", ErrUnusableScale, scale)
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.pixelsPerDIP = scale
	return nil
}

// perDIP answers the pixels to each DIP a window on monitor is sized with.
func (s *Service) perDIP(monitor placement.Monitor) float64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return sizingScale(s.pixelsPerDIP, monitor)
}

// sizingScale answers reported, the page's own scale, once the page has reported one; else the
// scale of monitor's DPI.
func sizingScale(reported float64, monitor placement.Monitor) float64 {
	if reported > 0 {
		return reported
	}
	return placement.PerDIPOf(monitor.DPI)
}

// New answers a service over ports with the first-run settings; Start loads the stored ones.
func New(ports Ports, layouts Layouts) *Service {
	return &Service{ports: ports, layouts: layouts, current: settings.Defaults()}
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
