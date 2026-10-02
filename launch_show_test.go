package main

import (
	"context"
	"testing"
)

// testMeasured is a measurement the page hands over at launch.
var testMeasured = measuredDTO{Size: "large", Style: "digital", Format: "24h", DateFormat: "day-month"}

// The page's sizing may land before it is ready: the ribbon is then shown the moment it is, with no
// fallback waited for.
func TestARibbonSizedBeforeItIsReadyShowsWhenReady(t *testing.T) {
	app, _, seen, _ := newTestApp(t)
	if err := app.SetPixelRatio(1); err != nil {
		t.Fatal(err)
	}
	if err := app.SetMeasured(testMeasured); err != nil {
		t.Fatal(err)
	}
	if seen.shown != 0 {
		t.Fatalf("shown %d times before the page was ready", seen.shown)
	}
	app.domReady(context.Background())
	if seen.shown != 1 || seen.sizePending != nil {
		t.Errorf("shown %d times with a fallback pending %v, want shown once at once", seen.shown, seen.sizePending != nil)
	}
}

// A page that never sizes the ribbon still shows it once the fallback falls due; a report that comes
// late does not show it again.
func TestARibbonThePageNeverSizesIsShownByTheFallbackOnce(t *testing.T) {
	app, _, seen, _ := newTestApp(t)
	app.domReady(context.Background())
	if seen.sizePending == nil {
		t.Fatal("no fallback waited for")
	}
	seen.sizePending()
	if seen.shown != 1 {
		t.Fatalf("shown %d times by the fallback, want once", seen.shown)
	}
	if err := app.SetPixelRatio(1); err != nil {
		t.Fatal(err)
	}
	if err := app.SetMeasured(testMeasured); err != nil {
		t.Fatal(err)
	}
	if seen.shown != 1 {
		t.Errorf("a late report showed the ribbon again: %d times", seen.shown)
	}
}

// A report the service refuses does not count as sizing the ribbon.
func TestARefusedReportDoesNotShowTheRibbon(t *testing.T) {
	app, service, seen, _ := newTestApp(t)
	app.domReady(context.Background())
	service.changeErr = errPlanted
	_ = app.SetPixelRatio(1)
	_ = app.SetMeasured(testMeasured)
	if seen.shown != 0 {
		t.Errorf("refused reports showed the ribbon %d times", seen.shown)
	}
}
