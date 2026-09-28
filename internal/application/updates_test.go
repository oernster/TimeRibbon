package application

import (
	"context"
	"errors"
	"testing"

	"github.com/oernster/timeribbon/internal/domain/settings"
)

// fakeReleases answers every check with info; with err instead when err is set.
type fakeReleases struct {
	info ReleaseInfo
	err  error
}

func (f fakeReleases) LatestRelease(context.Context) (ReleaseInfo, error) { return f.info, f.err }

const runningVersion = "2.0.0"

var newerRelease = ReleaseInfo{
	Version: "v2.1.0",
	PageURL: "https://example.test/release",
	Assets: []ReleaseAsset{
		{Name: "TimeRibbon.dmg", DownloadURL: "https://example.test/mac"},
		{Name: "TimeRibbonSetup.EXE", DownloadURL: "https://example.test/windows"},
	},
}

// updatingService answers a service over releases on Windows, loaded with skipped as the skipped
// release, together with the store it saves to.
func updatingService(t *testing.T, releases ReleaseSource, skipped string) (*Service, *fakeStore) {
	t.Helper()
	initial := settings.Defaults()
	initial.SkippedUpdate = skipped
	store := &fakeStore{loaded: Loaded{Settings: initial}}
	service := New(Ports{
		Store: store, Clock: fixedClock{}, Releases: releases,
		Build: Build{Version: runningVersion, Platform: PlatformWindows},
	}, testLayouts)
	if err := service.Start(); err != nil {
		t.Fatal(err)
	}
	return service, store
}

func TestIsNewerVersionComparesDottedIntegers(t *testing.T) {
	t.Parallel()
	cases := []struct {
		latest, current string
		newer           bool
	}{
		{"2.1.0", "2.0.0", true},
		{"v2.1.0", "2.0.0", true},
		{"V3.0.0", "2.9.9", true},
		{" 2.0.1 ", "2.0.0", true},
		{"2.0.0.1", "2.0.0", true},
		{"2.10.0", "2.9.0", true},
		{"2.0.0", "2.0.0", false},
		{"v2.0.0", "2.0.0", false},
		{"1.9.0", "2.0.0", false},
		{"2.0.0", "2.0.0.1", false},
		{"2.1.0-rc1", "2.0.0", false},
		{"banana", "2.0.0", false},
		{"2.1.0", "0.0.0-dev", false},
		{"", "2.0.0", false},
	}
	for _, c := range cases {
		if got := IsNewerVersion(c.latest, c.current); got != c.newer {
			t.Errorf("IsNewerVersion(%q, %q) = %v, want %v", c.latest, c.current, got, c.newer)
		}
	}
}

func TestEachSystemDownloadsItsOwnAsset(t *testing.T) {
	t.Parallel()
	for goos, want := range map[string]string{"windows": PlatformWindows, "darwin": PlatformMacOS, "linux": PlatformLinux, "freebsd": PlatformLinux} {
		if got := PlatformKeyFor(goos); got != want {
			t.Errorf("%s is %s, want %s", goos, got, want)
		}
	}
	for platform, want := range map[string]string{
		PlatformWindows: "https://example.test/windows", PlatformMacOS: "https://example.test/mac", PlatformLinux: "", "amiga": "",
	} {
		if got := SelectAssetURL(newerRelease.Assets, platform); got != want {
			t.Errorf("%s downloads %q, want %q", platform, got, want)
		}
	}
	if got := SelectAssetURL(nil, PlatformWindows); got != "" {
		t.Errorf("a release with no assets offered %q", got)
	}
}

// FR-509: a newer release is offered with this platform's download and its page.
func TestANewerReleaseIsOffered(t *testing.T) {
	t.Parallel()
	service, _ := updatingService(t, fakeReleases{info: newerRelease}, "")
	want := UpdateStatus{
		Current: runningVersion, Latest: "v2.1.0", UpdateAvailable: true,
		DownloadURL: "https://example.test/windows", PageURL: "https://example.test/release",
	}
	for _, manual := range []bool{false, true} {
		if got := service.CheckForUpdate(context.Background(), manual); got != want {
			t.Errorf("manual %v answered %+v", manual, got)
		}
	}
}

// FR-509: GitHub out of reach is no update and no latest version, whoever asked.
func TestAnUnreachableSourceOffersNothing(t *testing.T) {
	t.Parallel()
	service, _ := updatingService(t, fakeReleases{err: errors.New("offline")}, "")
	if got := service.CheckForUpdate(context.Background(), true); got != (UpdateStatus{Current: runningVersion}) {
		t.Errorf("answered %+v", got)
	}
}

// FR-509: the running version itself is seen but not offered.
func TestTheRunningVersionIsNotOffered(t *testing.T) {
	t.Parallel()
	same := ReleaseInfo{Version: "v" + runningVersion, PageURL: "https://example.test/release"}
	service, _ := updatingService(t, fakeReleases{info: same}, "")
	if got := service.CheckForUpdate(context.Background(), true); got.UpdateAvailable || got.Latest != same.Version {
		t.Errorf("answered %+v", got)
	}
}

// FR-509: a skipped release is never offered unasked; asking still offers it. A later release than
// the one skipped is offered either way.
func TestASkippedReleaseIsOfferedOnlyWhenAskedFor(t *testing.T) {
	t.Parallel()
	skipped, _ := updatingService(t, fakeReleases{info: newerRelease}, newerRelease.Version)
	if got := skipped.CheckForUpdate(context.Background(), false); got.UpdateAvailable || got.Latest != newerRelease.Version {
		t.Errorf("the automatic check answered %+v", got)
	}
	if got := skipped.CheckForUpdate(context.Background(), true); !got.UpdateAvailable {
		t.Errorf("the manual check answered %+v", got)
	}
	older, _ := updatingService(t, fakeReleases{info: newerRelease}, "v2.0.5")
	if got := older.CheckForUpdate(context.Background(), false); !got.UpdateAvailable {
		t.Errorf("a release later than the one skipped answered %+v", got)
	}
}

// FR-509: skipping keeps the version in the settings, saved.
func TestSkippingKeepsTheVersion(t *testing.T) {
	t.Parallel()
	service, store := updatingService(t, fakeReleases{info: newerRelease}, "")
	if err := service.SkipUpdate(newerRelease.Version); err != nil {
		t.Fatal(err)
	}
	if service.Settings().SkippedUpdate != newerRelease.Version || store.last(t).SkippedUpdate != newerRelease.Version {
		t.Errorf("kept %q, saved %q", service.Settings().SkippedUpdate, store.last(t).SkippedUpdate)
	}
	if got := service.CheckForUpdate(context.Background(), false); got.UpdateAvailable {
		t.Errorf("the skipped release was offered again: %+v", got)
	}
}
