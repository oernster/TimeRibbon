package application

// The update check (FR-509). The service asks the release source for the latest published release
// and compares it with the running version. A source that cannot be reached is never a fault the
// user hears about unasked: it answers a status with no latest version, which the automatic check
// keeps to itself and the manual check reports. A release the user skipped is seen but not offered,
// so the same version never prompts twice unasked. Ported from PigeonPost.

import (
	"context"
	"strconv"
	"strings"

	"github.com/oernster/timeribbon/internal/domain/settings"
)

// ReleaseAsset names one downloadable file attached to a release.
type ReleaseAsset struct {
	Name        string
	DownloadURL string
}

// ReleaseInfo is the latest published release as the release source reports it.
type ReleaseInfo struct {
	Version string
	PageURL string
	Assets  []ReleaseAsset
}

// ReleaseSource answers the latest published release. Only a published release that is neither a
// draft nor a prerelease is ever answered, so a tag pushed during development can never prompt.
type ReleaseSource interface {
	LatestRelease(ctx context.Context) (ReleaseInfo, error)
}

// Build is the build that is running: its version and the platform key naming which release asset
// it downloads.
type Build struct {
	Version  string
	Platform string
}

// UpdateStatus is the outcome of one update check. Latest is empty when the source could not be
// reached, when UpdateAvailable is always false. DownloadURL is this platform's asset; empty when the
// release carries none, when PageURL, the release page, stands in for it.
type UpdateStatus struct {
	Current         string
	Latest          string
	UpdateAvailable bool
	DownloadURL     string
	PageURL         string
}

// Platform keys naming which release asset the running system wants.
const (
	PlatformWindows = "windows"
	PlatformMacOS   = "macos"
	PlatformLinux   = "linux"
)

// versionTagPrefix is the optional prefix of a release tag, as in v2.0.0.
const versionTagPrefix = "v"

// versionSeparator separates a version's components.
const versionSeparator = "."

// assetSuffixes are each platform's asset filename endings, in the order they are preferred.
var assetSuffixes = map[string][]string{
	PlatformWindows: {".exe"},
	PlatformMacOS:   {".dmg"},
	PlatformLinux:   {".flatpak"},
}

// goosPlatforms maps each runtime.GOOS with a platform key of its own; any other is Linux.
var goosPlatforms = map[string]string{"windows": PlatformWindows, "darwin": PlatformMacOS}

// PlatformKeyFor answers the platform key of a runtime.GOOS value.
func PlatformKeyFor(goos string) string {
	if key, ok := goosPlatforms[goos]; ok {
		return key
	}
	return PlatformLinux
}

// SelectAssetURL answers the download address of the first asset for platform; empty when none is.
func SelectAssetURL(assets []ReleaseAsset, platform string) string {
	for _, suffix := range assetSuffixes[platform] {
		for _, asset := range assets {
			if strings.HasSuffix(strings.ToLower(asset.Name), suffix) {
				return asset.DownloadURL
			}
		}
	}
	return ""
}

// versionParts reads a dotted version with an optional leading v as integers; false when any
// component is not one.
func versionParts(version string) ([]int, bool) {
	text := strings.TrimSpace(version)
	if strings.HasPrefix(strings.ToLower(text), versionTagPrefix) {
		text = text[len(versionTagPrefix):]
	}
	fields := strings.Split(text, versionSeparator)
	parts := make([]int, 0, len(fields))
	for _, field := range fields {
		n, err := strconv.Atoi(field)
		if err != nil {
			return nil, false
		}
		parts = append(parts, n)
	}
	return parts, true
}

// IsNewerVersion reports whether latest is strictly newer than current. A version that is not dotted
// integers is never newer, so a malformed tag cannot prompt.
func IsNewerVersion(latest, current string) bool {
	latestParts, ok := versionParts(latest)
	if !ok {
		return false
	}
	currentParts, ok := versionParts(current)
	if !ok {
		return false
	}
	for i := 0; i < len(latestParts) && i < len(currentParts); i++ {
		if latestParts[i] != currentParts[i] {
			return latestParts[i] > currentParts[i]
		}
	}
	return len(latestParts) > len(currentParts)
}

// CheckForUpdate answers the update status of the running build. The automatic check passes manual
// false and is never offered the release the user skipped; a manual check is asked for, so it
// ignores the skip.
func (s *Service) CheckForUpdate(ctx context.Context, manual bool) UpdateStatus {
	build := s.ports.Build
	info, err := s.ports.Releases.LatestRelease(ctx)
	if err != nil {
		return UpdateStatus{Current: build.Version}
	}
	skipped := s.Settings().SkippedUpdate
	return UpdateStatus{
		Current:         build.Version,
		Latest:          info.Version,
		UpdateAvailable: IsNewerVersion(info.Version, build.Version) && (manual || info.Version != skipped),
		DownloadURL:     SelectAssetURL(info.Assets, build.Platform),
		PageURL:         info.PageURL,
	}
}

// SkipUpdate keeps version as the release the automatic check never offers again (FR-509).
func (s *Service) SkipUpdate(version string) error {
	return choose(s, version, func(c *settings.Settings) *string { return &c.SkippedUpdate })
}
