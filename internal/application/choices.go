package application

import (
	"fmt"

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

// SetColour chooses the colour scheme every clock is drawn in (FR-611).
func (s *Service) SetColour(colour settings.Colour) error {
	return choose(s, colour, func(c *settings.Settings) *settings.Colour { return &c.Colour })
}

// SetFormat chooses 12-hour or 24-hour time (FR-206).
func (s *Service) SetFormat(format clock.Format) error {
	return choose(s, format, func(c *settings.Settings) *clock.Format { return &c.Format })
}

// SetDateFormat chooses how every date is written (FR-612).
func (s *Service) SetDateFormat(dateFormat clock.DateFormat) error {
	return choose(s, dateFormat, func(c *settings.Settings) *clock.DateFormat { return &c.DateFormat })
}

// SetOrientation chooses horizontal or vertical (FR-103). The window then goes to the
// orientation's home edge (FR-409).
func (s *Service) SetOrientation(orientation settings.Orientation) error {
	return choose(s, orientation, func(c *settings.Settings) *settings.Orientation { return &c.Orientation })
}

// SetTheme chooses system, light or dark (FR-606).
func (s *Service) SetTheme(theme settings.Theme) error {
	return choose(s, theme, func(c *settings.Settings) *settings.Theme { return &c.Theme })
}

// SetAlwaysOnTop turns Always on Top on or off (FR-505).
func (s *Service) SetAlwaysOnTop(on bool) error {
	return choose(s, on, func(c *settings.Settings) *bool { return &c.AlwaysOnTop })
}

// choose sets the field field picks to value and saves; a value the setting does not offer, which
// normalising would replace, is refused and nothing changes (FR-602).
func choose[T comparable](s *Service, value T, field func(*settings.Settings) *T) error {
	return s.change(func(current settings.Settings) (settings.Settings, error) {
		next := current.Normalised()
		*field(&next) = value
		normalised := next.Normalised()
		if *field(&normalised) != value {
			return current, fmt.Errorf("%w: %v", ErrUnknownChoice, value)
		}
		return next, nil
	})
}

// StartWithWindows answers whether the Start with Windows value is present. Windows holds the
// answer, not the settings file, so the setting and setup cannot disagree (FR-605, FR-805).
func (s *Service) StartWithWindows() (bool, error) {
	return s.ports.Startup.Enabled()
}

// SetStartWithWindows writes or removes the Start with Windows value (FR-605).
func (s *Service) SetStartWithWindows(on bool) error {
	if on {
		return s.ports.Startup.Enable()
	}
	return s.ports.Startup.Disable()
}
