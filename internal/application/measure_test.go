package application

import (
	"errors"
	"slices"
	"testing"

	"github.com/oernster/timeribbon/internal/domain/clock"
	"github.com/oernster/timeribbon/internal/domain/settings"
)

// measuredFor answers a measurement of width taken under current's choices.
func measuredFor(current settings.Settings, width int) Measured {
	current = current.Normalised()
	return Measured{Size: current.Size, Style: current.Style, Format: current.Format, DateFormat: current.DateFormat, CellWidth: width}
}

// FR-620: the samples are written in the chosen formats.
func TestTheTextSamplesFollowTheFormats(t *testing.T) {
	t.Parallel()
	chosen := clocks(1)
	chosen.Format, chosen.DateFormat = clock.TwelveHour, clock.DayMonthYear
	times, dates := newRig(t, chosen).service.TextSamples()
	if !slices.Contains(times, "11:59 PM") || !slices.Contains(dates, "Wed 30/09/2026") {
		t.Errorf("got %d times, %d dates, starting %q and %q", len(times), len(dates), times[0], dates[0])
	}
}

// FR-620: a measured width wider than the size's own widens every clock cell, in the snapshot the
// page draws from and in the window alike; a narrower one leaves the size's own.
func TestAMeasuredWidthWidensTheCells(t *testing.T) {
	t.Parallel()
	wider := testLayout.Digital.Width + testLayout.Padding
	for _, orientation := range []settings.Orientation{settings.Horizontal, settings.Vertical} {
		initial := clocks(2)
		initial.Orientation = orientation
		r := newRig(t, initial)
		if err := r.service.SetMeasured(measuredFor(initial, wider)); err != nil {
			t.Fatal(err)
		}
		if got := r.service.Snapshot().Layout.Digital.Width; got != wider {
			t.Errorf("%s: snapshot cell %d, want %d", orientation, got, wider)
		}
		got, err := r.service.Launch()
		if err != nil {
			t.Fatal(err)
		}
		want, across := 2*wider+2*testLayout.Padding, got.Size.Width
		if orientation == settings.Vertical {
			want = wider + 2*testLayout.Padding
		}
		if across != want {
			t.Errorf("%s: window %d wide, want %d", orientation, across, want)
		}
		if err := r.service.SetMeasured(measuredFor(initial, testLayout.Digital.Width-1)); err != nil {
			t.Fatal(err)
		}
		if got := r.service.Snapshot().Layout.Digital.Width; got != testLayout.Digital.Width {
			t.Errorf("%s: a narrower measurement gave %d", orientation, got)
		}
	}
}

// FR-620: a measurement taken under other choices widens nothing until the page measures again;
// an analogue one widens the analogue cell alone; a width below zero is refused.
func TestAMeasurementCountsOnlyForItsOwnChoices(t *testing.T) {
	t.Parallel()
	initial := clocks(2)
	r := newRig(t, initial)
	wider := testLayout.Digital.Width + testLayout.Padding
	for _, other := range []func(*Measured){
		func(m *Measured) { m.Size = settings.Small },
		func(m *Measured) { m.Style = settings.Analogue },
		func(m *Measured) { m.Format = clock.TwelveHour },
		func(m *Measured) { m.DateFormat = clock.YearMonthDay },
	} {
		measured := measuredFor(initial, wider)
		other(&measured)
		if err := r.service.SetMeasured(measured); err != nil {
			t.Fatal(err)
		}
		if got := r.service.Snapshot().Layout.Digital.Width; got != testLayout.Digital.Width {
			t.Errorf("%+v widened the cell to %d", measured, got)
		}
	}
	analogue := initial
	analogue.Style = settings.Analogue
	a := newRig(t, analogue)
	if err := a.service.SetMeasured(measuredFor(analogue, wider)); err != nil {
		t.Fatal(err)
	}
	if got := a.service.Snapshot().Layout; got.Analogue.Width != wider || got.Digital.Width != testLayout.Digital.Width {
		t.Errorf("analogue: %+v", got)
	}
	if err := r.service.SetMeasured(measuredFor(initial, -1)); !errors.Is(err, ErrNegativeLength) {
		t.Errorf("a negative width: %v", err)
	}
}
