package main

// The wire between Go and the page. Each type here is stated a second time in
// frontend/src/wire.ts; a structural test compares the two, since the type checker sees only the
// TypeScript and the marshaller sees only these.

import (
	"github.com/oernster/timestrip/internal/application"
	"github.com/oernster/timestrip/internal/domain/placement"
)

// sizeDTO is a width and a height in DIP.
type sizeDTO struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// layoutDTO is the cell geometry the page draws with.
type layoutDTO struct {
	Digital  sizeDTO `json:"digital"`
	Analogue sizeDTO `json:"analogue"`
	Prompt   sizeDTO `json:"prompt"`
	Padding  int     `json:"padding"`
}

// cellDTO is one cell of the strip.
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

// snapshotDTO is everything the strip draws.
type snapshotDTO struct {
	Cells         []cellDTO `json:"cells"`
	Style         string    `json:"style"`
	Format        string    `json:"format"`
	Orientation   string    `json:"orientation"`
	Theme         string    `json:"theme"`
	AlwaysOnTop   bool      `json:"alwaysOnTop"`
	Layout        layoutDTO `json:"layout"`
	RefreshInMs   int64     `json:"refreshInMs"`
	Notices       []string  `json:"notices"`
	Scrolls       bool      `json:"scrolls"`
	DragThreshold sizeDTO   `json:"dragThreshold"`
}

// placeDTO is one entry of the place search.
type placeDTO struct {
	Zone    string `json:"zone"`
	Label   string `json:"label"`
	Country string `json:"country"`
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
		Cells: cells, Style: string(s.Style), Format: string(s.Format), Orientation: string(s.Orientation),
		Theme: string(s.Theme), AlwaysOnTop: s.AlwaysOnTop,
		Layout:        layoutDTO{Digital: sizeOf(s.Layout.Digital), Analogue: sizeOf(s.Layout.Analogue), Prompt: sizeOf(s.Layout.Prompt), Padding: s.Layout.Padding},
		RefreshInMs:   s.NextRefresh.Sub(s.Now).Milliseconds(),
		Notices:       notices,
		Scrolls:       scrolls,
		DragThreshold: sizeOf(threshold),
	}
}

func placesOf(places []application.Place) []placeDTO {
	out := make([]placeDTO, 0, len(places))
	for _, p := range places {
		out = append(out, placeDTO{Zone: p.Zone, Label: p.Label, Country: p.Country})
	}
	return out
}
