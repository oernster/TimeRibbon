package main

import (
	"slices"
	"testing"

	"github.com/oernster/timeribbon/internal/application"
	"github.com/oernster/timeribbon/internal/domain/clock"
	"github.com/oernster/timeribbon/internal/domain/settings"
)

// FR-620: the page is handed the samples whole and its measurement reaches the service in the
// service's own terms; once it is applied, the window hears the page has been measured.
func TestTheMeasurementRoundTrip(t *testing.T) {
	app, service, control := newTestApp(t)
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
	if !slices.Equal(control.calls, []string{"Refitted", "PageMeasured"}) {
		t.Errorf("the window was asked %v, want the ribbon fitted then the page counted measured", control.calls)
	}
}

// A measurement the service refused does not count as applied, so the launched ribbon waits on.
func TestARefusedMeasurementIsNotCountedApplied(t *testing.T) {
	app, service, control := newTestApp(t)
	service.changeErr = errPlanted
	_ = app.SetMeasured(measuredDTO{CellWidth: 181})
	if slices.Contains(control.calls, "PageMeasured") {
		t.Error("a refused measurement counted the page measured")
	}
}
