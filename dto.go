package main

// The wire between Go and the page. Each type here is stated a second time in
// frontend/src/wire.ts; a structural test compares the two, since the type checker sees only the
// TypeScript and the marshaller sees only these.

import (
	"github.com/oernster/timeribbon/internal/application"
	"github.com/oernster/timeribbon/internal/product"
	"github.com/oernster/timeribbon/ribbonkit/application/menus"
	"github.com/oernster/timeribbon/ribbonkit/application/release"
	"github.com/oernster/timeribbon/ribbonkit/domain/placement"
)

// sizeDTO is a width and a height in DIP.
type sizeDTO struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// layoutDTO is the cell geometry the page draws with.
type layoutDTO struct {
	Digital    sizeDTO `json:"digital"`
	Analogue   sizeDTO `json:"analogue"`
	Prompt     sizeDTO `json:"prompt"`
	Padding    int     `json:"padding"`
	HandleLane int     `json:"handleLane"`
}

// textSamplesDTO is every time and date a cell can show under the current formats (FR-620).
type textSamplesDTO struct {
	Times []string `json:"times"`
	Dates []string `json:"dates"`
}

// measuredDTO is the cell width the page measured its widest text to need, with the choices it
// measured under (FR-620).
type measuredDTO struct {
	Size       string `json:"size"`
	Style      string `json:"style"`
	Format     string `json:"format"`
	DateFormat string `json:"dateFormat"`
	CellWidth  int    `json:"cellWidth"`
}

// cellDTO is one cell of the ribbon.
type cellDTO struct {
	ID          string  `json:"id"`
	Label       string  `json:"label"`
	Zone        string  `json:"zone"`
	ZoneMark    string  `json:"zoneMark"`
	Time        string  `json:"time"`
	Date        string  `json:"date"`
	HourAngle   float64 `json:"hourAngle"`
	MinuteAngle float64 `json:"minuteAngle"`
	Problem     string  `json:"problem"`
}

// snapshotDTO is everything the ribbon draws.
type snapshotDTO struct {
	Cells         []cellDTO `json:"cells"`
	Style         string    `json:"style"`
	Size          string    `json:"size"`
	Colour        string    `json:"colour"`
	Format        string    `json:"format"`
	DateFormat    string    `json:"dateFormat"`
	Orientation   string    `json:"orientation"`
	Theme         string    `json:"theme"`
	AlwaysOnTop   bool      `json:"alwaysOnTop"`
	Opacity       int       `json:"opacity"`
	MinOpacity    int       `json:"minOpacity"`
	Scale         float64   `json:"scale"`
	MinScale      int       `json:"minScale"`
	MaxScale      int       `json:"maxScale"`
	Layout        layoutDTO `json:"layout"`
	RefreshInMs   int64     `json:"refreshInMs"`
	Notices       []string  `json:"notices"`
	Scrolls       bool      `json:"scrolls"`
	DragThreshold sizeDTO   `json:"dragThreshold"`
	StartLabel    string    `json:"startLabel"`
	// Collapsed is true while the window is an unpinned ribbon's tab (FR-614).
	Collapsed bool      `json:"collapsed"`
	SunMap    sunMapDTO `json:"sunMap"`
	// Choices are the menus' choices, which Settings offers as well (FR-624).
	Choices []choiceDTO `json:"choices"`
}

// choiceDTO is one of the menus' choices as Settings draws it (FR-624): either a group of Children
// or one item whose Action the page hands back to Choose.
type choiceDTO struct {
	Action    string      `json:"action"`
	Label     string      `json:"label"`
	Checkable bool        `json:"checkable"`
	Checked   bool        `json:"checked"`
	Children  []choiceDTO `json:"children"`
}

// choicesOf answers the wire form of menu items, every list present so the page never meets null.
func choicesOf(items []menus.Item) []choiceDTO {
	out := make([]choiceDTO, 0, len(items))
	for _, item := range items {
		out = append(out, choiceDTO{
			Action: string(item.Action), Label: item.Label, Checkable: item.Checkable, Checked: item.Checked,
			Children: choicesOf(item.Children),
		})
	}
	return out
}

// boxDTO is a rectangle inside the window, in the page's units.
type boxDTO struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// markDTO is one clock's place on the sun map (FR-908).
type markDTO struct {
	Label     string  `json:"label"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// sunMapDTO is what the sun map draws and where (FR-901 to FR-910). Side is where the map and the
// pull out's handle go, empty while the map is off; Shown is whether the map is drawn now, in Map,
// beside the ribbon in Ribbon. Latitude and Longitude are the subsolar point.
type sunMapDTO struct {
	On        bool      `json:"on"`
	PullOut   bool      `json:"pullOut"`
	Side      string    `json:"side"`
	Shown     bool      `json:"shown"`
	Ribbon    boxDTO    `json:"ribbon"`
	Map       boxDTO    `json:"map"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Marks     []markDTO `json:"marks"`
}

// boxOf answers r, a rectangle in window pixels, in the page's units: divided by perDIP, the window
// pixels to each of them. The page then draws the box as it is, whatever size the window has at that
// moment; an opening ribbon is drawn while the window is still its tab (FR-615). Before the page has
// reported its ratio perDIP is zero and r goes as it is, one pixel to a unit.
func boxOf(r placement.Rect, perDIP float64) boxDTO {
	if perDIP == 0 {
		perDIP = 1
	}
	return boxDTO{
		X: float64(r.Left) / perDIP, Y: float64(r.Top) / perDIP,
		Width: float64(r.Width()) / perDIP, Height: float64(r.Height()) / perDIP,
	}
}

// sunMapOf answers the wire form of the sun map's content; where it is drawn is the facade's.
func sunMapOf(m application.SunMap) sunMapDTO {
	marks := make([]markDTO, 0, len(m.Marks))
	for _, mark := range m.Marks {
		marks = append(marks, markDTO{Label: mark.Label, Latitude: mark.At.Latitude, Longitude: mark.At.Longitude})
	}
	return sunMapDTO{On: m.On, PullOut: m.PullOut, Latitude: m.Subsolar.Latitude, Longitude: m.Subsolar.Longitude, Marks: marks}
}

// placeDTO is one entry of the place search.
type placeDTO struct {
	Zone    string `json:"zone"`
	Label   string `json:"label"`
	Country string `json:"country"`
}

// aboutDTO is what the About panel shows (FR-607).
type aboutDTO struct {
	Name      string      `json:"name"`
	Version   string      `json:"version"`
	Author    string      `json:"author"`
	Copyright string      `json:"copyright"`
	Credits   []creditDTO `json:"credits"`
}

// creditDTO is one component the application ships.
type creditDTO struct {
	Name    string `json:"name"`
	Licence string `json:"licence"`
	Role    string `json:"role"`
}

// updateDTO is what the update panel shows (FR-509). Latest is empty when GitHub could not be
// reached. The addresses stay in Go: the page asks Go to open what it offered, never names one.
type updateDTO struct {
	Current         string `json:"current"`
	Latest          string `json:"latest"`
	UpdateAvailable bool   `json:"updateAvailable"`
}

func updateOf(status release.Status) updateDTO {
	return updateDTO{Current: status.Current, Latest: status.Latest, UpdateAvailable: status.Available}
}

func sizeOf(size placement.Size) sizeDTO { return sizeDTO{Width: size.Width, Height: size.Height} }

// snapshotOf answers the wire form of a snapshot.
func snapshotOf(s application.Snapshot, scrolls bool, threshold placement.Size) snapshotDTO {
	cells := make([]cellDTO, 0, len(s.Cells))
	for _, c := range s.Cells {
		cells = append(cells, cellDTO{
			ID: c.ID, Label: c.Label, Zone: c.Zone, ZoneMark: c.ZoneMark, Time: c.Time, Date: c.Date,
			HourAngle: c.HourAngle, MinuteAngle: c.MinuteAngle, Problem: c.Problem,
		})
	}
	notices := s.Notices
	if notices == nil {
		notices = []string{}
	}
	return snapshotDTO{
		Cells: cells, Style: string(s.Style), Size: string(s.Size), Colour: string(s.Colour), Format: string(s.Format), DateFormat: string(s.DateFormat), Orientation: string(s.Orientation),
		Theme: string(s.Theme), AlwaysOnTop: s.AlwaysOnTop, Opacity: s.Opacity, MinOpacity: s.MinOpacity,
		Scale: s.Scale, MinScale: s.MinScale, MaxScale: s.MaxScale,
		Layout:        layoutDTO{Digital: sizeOf(s.Layout.Digital), Analogue: sizeOf(s.Layout.Analogue), Prompt: sizeOf(s.Layout.Prompt), Padding: s.Layout.Padding, HandleLane: s.Layout.HandleLane},
		RefreshInMs:   s.NextRefresh.Sub(s.Now).Milliseconds(),
		Notices:       notices,
		Scrolls:       scrolls,
		DragThreshold: sizeOf(threshold),
		StartLabel:    product.StartAtSignIn,
		SunMap:        sunMapOf(s.SunMap),
	}
}

func placesOf(places []application.Place) []placeDTO {
	out := make([]placeDTO, 0, len(places))
	for _, p := range places {
		out = append(out, placeDTO{Zone: p.Zone, Label: p.Label, Country: p.Country})
	}
	return out
}
