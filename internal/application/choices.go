package application

import (
	"fmt"

	"github.com/oernster/ribbonkit/domain/localtime"
	"github.com/oernster/ribbonkit/domain/ribbon"
	"github.com/oernster/timeribbon/internal/domain/clock"
	"github.com/oernster/timeribbon/internal/domain/settings"
)

// SetStyle chooses digital or analogue presentation (FR-601 to FR-604).
func (s *Service) SetStyle(style settings.Style) error {
	return choose(s, style, func(c *settings.Settings) *settings.Style { return &c.Style })
}

// SetSize chooses large or small cells (FR-610). Arranging the window afterwards fits it to them.
func (s *Service) SetSize(size settings.Size) error {
	return choose(s, size, func(c *settings.Settings) *settings.Size { return &c.Size })
}

// SetFormat chooses 12-hour or 24-hour time (FR-206).
func (s *Service) SetFormat(format localtime.Format) error {
	return choose(s, format, func(c *settings.Settings) *localtime.Format { return &c.Format })
}

// SetDateFormat chooses how every date is written (FR-612).
func (s *Service) SetDateFormat(dateFormat clock.DateFormat) error {
	return choose(s, dateFormat, func(c *settings.Settings) *clock.DateFormat { return &c.DateFormat })
}

// choose sets the field field picks to value and saves; a value the setting does not offer, which
// normalising would replace, is refused and nothing changes (FR-602).
func choose[T comparable](s *Service, value T, field func(*settings.Settings) *T) error {
	return s.change(func(current settings.Settings) (settings.Settings, error) {
		next := current.Normalised()
		*field(&next) = value
		normalised := next.Normalised()
		if *field(&normalised) != value {
			return current, fmt.Errorf("%w: %v", ribbon.ErrUnknownChoice, value)
		}
		return next, nil
	})
}
