package application

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/oernster/timeribbon/internal/domain/clock"
	"github.com/oernster/timeribbon/internal/domain/settings"
)

// FR-704: the store's notice reaches the ribbon.
func TestALoadNoticeIsShownUntilDismissed(t *testing.T) {
	t.Parallel()
	store := &fakeStore{loaded: Loaded{Settings: settings.Defaults(), Notice: "kept aside"}}
	service := New(Ports{Store: store, Clock: fixedClock{}}, testLayouts)
	if err := service.Start(); err != nil {
		t.Fatal(err)
	}
	if got := service.Snapshot().Notices; !slices.Equal(got, []string{"kept aside"}) {
		t.Errorf("got %v", got)
	}
	service.DismissNotices()
	if got := service.Snapshot().Notices; len(got) != 0 {
		t.Errorf("after dismissing: %v", got)
	}
}

func TestAFaultReadingSettingsKeepsTheDefaultsAndSaysSo(t *testing.T) {
	t.Parallel()
	store := &fakeStore{loadErr: errPlanted}
	service := New(Ports{Store: store, Clock: fixedClock{}}, testLayouts)
	if err := service.Start(); !errors.Is(err, errPlanted) {
		t.Errorf("Start answered %v", err)
	}
	if got := service.Settings(); got.Style != settings.Digital || len(got.Clocks) != 0 {
		t.Errorf("settings %+v", got)
	}
	notices := service.Snapshot().Notices
	if len(notices) != 1 || !strings.Contains(notices[0], errPlanted.Error()) {
		t.Errorf("notices %v", notices)
	}
}

// FR-707.
func TestWriteFailureIsReportedAndCleared(t *testing.T) {
	t.Parallel()
	r := newRig(t, settings.Defaults())
	r.store.saveErr = errPlanted
	if err := r.service.SetFormat(clock.TwelveHour); !errors.Is(err, errPlanted) {
		t.Errorf("got %v", err)
	}
	snapshot := r.service.Snapshot()
	if snapshot.Format != clock.TwelveHour {
		t.Error("the change was not kept in effect")
	}
	if len(snapshot.Notices) != 1 || snapshot.Notices[0] != saveFailedPrefix+errPlanted.Error() {
		t.Errorf("notices %v", snapshot.Notices)
	}
	r.store.saveErr = nil
	if err := r.service.SetFormat(clock.TwentyFourHour); err != nil {
		t.Fatal(err)
	}
	if got := r.service.Snapshot().Notices; len(got) != 0 {
		t.Errorf("a later save that succeeds clears the notice: %v", got)
	}
}

// FR-602.
func TestChangingASettingPersistsIt(t *testing.T) {
	t.Parallel()
	r := newRig(t, settings.Defaults())
	steps := []error{
		r.service.SetStyle(settings.Analogue),
		r.service.SetFormat(clock.TwelveHour),
		r.service.SetDateFormat(clock.DayMonthYear),
		r.service.SetOrientation(settings.Vertical),
		r.service.SetTheme(settings.Dark),
		r.service.SetAlwaysOnTop(true),
	}
	for index, err := range steps {
		if err != nil {
			t.Fatalf("step %d: %v", index, err)
		}
	}
	saved := r.store.last(t)
	if saved.Style != settings.Analogue || saved.Format != clock.TwelveHour || saved.DateFormat != clock.DayMonthYear ||
		saved.Orientation != settings.Vertical || saved.Theme != settings.Dark || !saved.AlwaysOnTop {
		t.Errorf("saved %+v", saved)
	}
}

func TestAValueASettingDoesNotOfferIsRefused(t *testing.T) {
	t.Parallel()
	r := newRig(t, settings.Defaults())
	for name, err := range map[string]error{
		"style":       r.service.SetStyle("sundial"),
		"format":      r.service.SetFormat("36h"),
		"date format": r.service.SetDateFormat("stardate"),
		"orientation": r.service.SetOrientation("diagonal"),
		"theme":       r.service.SetTheme("sepia"),
	} {
		if !errors.Is(err, ErrUnknownChoice) {
			t.Errorf("%s: got %v", name, err)
		}
	}
	if len(r.store.saved) != 0 || r.service.Settings().Style != settings.Digital {
		t.Error("a refused value changed something")
	}
}

// FR-605: Windows holds the answer.
func TestStartWithWindowsPassesThroughToTheEntry(t *testing.T) {
	t.Parallel()
	r := newRig(t, settings.Defaults())
	if err := r.service.SetStartWithWindows(true); err != nil {
		t.Fatal(err)
	}
	if on, err := r.service.StartWithWindows(); err != nil || !on {
		t.Errorf("after enabling: %v %v", on, err)
	}
	if err := r.service.SetStartWithWindows(false); err != nil {
		t.Fatal(err)
	}
	if on, _ := r.service.StartWithWindows(); on {
		t.Error("still enabled after disabling")
	}
	if len(r.store.saved) != 0 {
		t.Error("the settings file was written for a value Windows holds")
	}
}
