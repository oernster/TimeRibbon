package application

// The update check (FR-509). Its rules are ribbonkit's release package; the service supplies the
// release the user skipped, which it keeps in the settings.

import (
	"context"

	"github.com/oernster/timeribbon/internal/domain/settings"
	"github.com/oernster/timeribbon/ribbonkit/application/release"
)

// CheckForUpdate answers the update status of the running build. The automatic check passes manual
// false and is never offered the release the user skipped; a manual check is asked for, so it
// ignores the skip.
func (s *Service) CheckForUpdate(ctx context.Context, manual bool) release.Status {
	return release.Check(ctx, s.ports.Releases, s.ports.Build, s.Settings().SkippedUpdate, manual)
}

// SkipUpdate keeps version as the release the automatic check never offers again (FR-509).
func (s *Service) SkipUpdate(version string) error {
	return choose(s, version, func(c *settings.Settings) *string { return &c.SkippedUpdate })
}
