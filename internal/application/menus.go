package application

import (
	"github.com/oernster/ribbonkit/application/menus"
)

// TimeRibbon's own actions; every ribbon's are ribbonkit's menus package's, with their items and
// words. The words shown for TimeRibbon's own live here alone.
const (
	ActionAddClock menus.Action = "add-clock"
	ActionSunMap   menus.Action = "sun-map"
)

// Item words of TimeRibbon's own, one home each.
const (
	labelAddClock = "Add clock"
	labelSunMap   = "Sun map"
)

// TrayMenu answers the tray menu for a ribbon that is or is not visible (FR-502).
func (s *Service) TrayMenu(visible bool) []menus.Item {
	choices := s.Settings().Choices
	return []menus.Item{
		menus.Visibility(visible), addClockItem(), menus.SettingsItem(), s.styleItem(), menus.ColourItem(choices.Colour),
		menus.OrientationItem(choices.Orientation), menus.PositionItem(choices.Orientation),
		menus.AlwaysOnTopItem(choices.AlwaysOnTop), menus.PinItem(choices.Pinned), s.sunMapItem(), menus.HelpItem(),
		menus.ExitItem(),
	}
}

// ContextMenu answers the menu the ribbon offers when right-clicked (FR-108).
func (s *Service) ContextMenu() []menus.Item {
	choices := s.Settings().Choices
	return []menus.Item{
		addClockItem(), menus.SettingsItem(), s.styleItem(), menus.ColourItem(choices.Colour),
		menus.OrientationItem(choices.Orientation), menus.PositionItem(choices.Orientation),
		menus.AlwaysOnTopItem(choices.AlwaysOnTop), menus.PinItem(choices.Pinned), s.sunMapItem(), menus.HelpItem(),
		menus.HideItem(), menus.ExitItem(),
	}
}

// SettingsChoices answers every choice the menus offer, for Settings to offer as well (FR-624): the
// same items both menus hold, so their words and ticks have one home. What is left of the menus is
// commands (show or hide, Add clock, Settings, Help, Exit), which choose nothing.
func (s *Service) SettingsChoices() []menus.Item {
	choices := s.Settings().Choices
	return []menus.Item{
		s.styleItem(), menus.ColourItem(choices.Colour), menus.OrientationItem(choices.Orientation),
		menus.PositionItem(choices.Orientation), menus.AlwaysOnTopItem(choices.AlwaysOnTop),
		menus.PinItem(choices.Pinned), s.sunMapItem(),
	}
}

// CloseRequested answers what a request to close the ribbon does, such as Alt+F4: it hides the
// ribbon and the application keeps running; only Exit ends it (FR-504, FR-507).
func (s *Service) CloseRequested() menus.Action {
	return menus.Hide
}

func addClockItem() menus.Item { return menus.Item{Action: ActionAddClock, Label: labelAddClock} }
