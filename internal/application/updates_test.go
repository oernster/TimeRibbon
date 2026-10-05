package application

import (
	"context"
	"testing"

	"github.com/oernster/timeribbon/internal/domain/settings"
	"github.com/oernster/timeribbon/ribbonkit/application/release"
)

// fakeReleases answers every check with info.
type fakeReleases struct{ info release.Info }

func (f fakeReleases) LatestRelease(context.Context) (release.Info, error) { return f.info, nil }

var newerRelease = release.Info{Version: "v2.1.0", PageURL: "https://example.test/release"}

// FR-509: skipping keeps the version in the settings, saved, and the service hands the kept skip to
// the check, so the automatic check no longer offers it. The check's own rules are ribbonkit's
// release package's tests.
func TestSkippingKeepsTheVersion(t *testing.T) {
	t.Parallel()
	store := &fakeStore{loaded: Loaded{Settings: settings.Defaults()}}
	service := New(Ports{
		Store: store, Clock: fixedClock{}, Releases: fakeReleases{info: newerRelease},
		Build: release.Build{Version: "2.0.0", Platform: release.PlatformWindows},
	}, testLayouts)
	if err := service.Start(); err != nil {
		t.Fatal(err)
	}
	if got := service.CheckForUpdate(context.Background(), false); !got.Available {
		t.Fatalf("before skipping, the newer release was not offered: %+v", got)
	}
	if err := service.SkipUpdate(newerRelease.Version); err != nil {
		t.Fatal(err)
	}
	if service.Settings().SkippedUpdate != newerRelease.Version || store.last(t).SkippedUpdate != newerRelease.Version {
		t.Errorf("kept %q, saved %q", service.Settings().SkippedUpdate, store.last(t).SkippedUpdate)
	}
	if got := service.CheckForUpdate(context.Background(), false); got.Available {
		t.Errorf("the skipped release was offered again: %+v", got)
	}
}
