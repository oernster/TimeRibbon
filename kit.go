package main

// What TimeRibbon hands ribbonkit's window: the page it serves, the terms Help shows and its
// application service as the window asks for it.

import (
	"embed"

	"github.com/oernster/timeribbon/internal/application"
	"github.com/oernster/timeribbon/ribbonkit/domain/ribbon"
)

// assets is the built page the window serves.
//
//go:embed all:frontend/dist
var assets embed.FS

// licenceText is the LICENSE file itself, so the Licence panel cannot disagree with what was built
// (FR-608).
//
//go:embed LICENSE
var licenceText string

// kitService is TimeRibbon's application service as the window asks for it: the ribbon's choices and
// the pull out read out of TimeRibbon's settings, every other call the service's own.
type kitService struct {
	*application.Service
}

// Choices answers the ribbon's own choices out of TimeRibbon's settings.
func (k kitService) Choices() ribbon.Choices { return k.Settings().Choices }

// PullOut answers whether the sun map is pulled out beside a vertical ribbon (FR-903).
func (k kitService) PullOut() bool { return k.Settings().PullOut }
