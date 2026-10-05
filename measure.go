package main

import (
	"github.com/oernster/timeribbon/internal/application"
	"github.com/oernster/timeribbon/internal/domain/clock"
	"github.com/oernster/timeribbon/internal/domain/settings"
)

// TextSamples answers every time and date a cell can show, for the page to measure (FR-620).
func (a *App) TextSamples() textSamplesDTO {
	times, dates := a.service.TextSamples()
	return textSamplesDTO{Times: times, Dates: dates}
}

// SetMeasured takes the cell width the page measured its widest time and date to need in the font
// it really draws with, then fits the ribbon to it (FR-620).
func (a *App) SetMeasured(measured measuredDTO) error {
	err := a.control.Refitted(a.service.SetMeasured(application.Measured{
		Size:       settings.Size(measured.Size),
		Style:      settings.Style(measured.Style),
		Format:     clock.Format(measured.Format),
		DateFormat: clock.DateFormat(measured.DateFormat),
		CellWidth:  measured.CellWidth,
	}))
	if err == nil {
		a.control.PageMeasured()
	}
	return err
}
