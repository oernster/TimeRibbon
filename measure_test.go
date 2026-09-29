package main

import (
	"testing"

	"github.com/oernster/timeribbon/internal/application"
	"github.com/oernster/timeribbon/internal/domain/clock"
	"github.com/oernster/timeribbon/internal/domain/settings"
)

// FR-620: the page is handed the samples whole and its measurement reaches the service in the
// service's own terms.
func TestTheMeasurementRoundTrip(t *testing.T) {
	app, service, _, _ := newTestApp(t)
	samples := app.TextSamples()
	if len(samples.Times) != 1 || samples.Times[0] != "23:59" || len(samples.Dates) != 1 || samples.Dates[0] != "Wednesday, 30 September" {
		t.Errorf("samples %+v", samples)
	}
	if err := app.SetMeasured(measuredDTO{Size: "small", Style: "analogue", Format: "12h", DateFormat: "ymd", CellWidth: 181}); err != nil {
		t.Fatal(err)
	}
	want := application.Measured{Size: settings.Small, Style: settings.Analogue, Format: clock.TwelveHour, DateFormat: clock.YearMonthDay, CellWidth: 181}
	if service.measured != want {
		t.Errorf("service was handed %+v, want %+v", service.measured, want)
	}
}
