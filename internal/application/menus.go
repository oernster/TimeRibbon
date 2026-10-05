package application

import (
	"github.com/oernster/timeribbon/ribbonkit/application/menus"
	"github.com/oernster/timeribbon/ribbonkit/domain/ribbon"
)

// TimeRibbon's own actions; every ribbon's are ribbonkit's menus package's. The words shown for each
// live here alone.
const (
	ActionAddClock menus.Action = "add-clock"
	ActionSunMap   menus.Action = "sun-map"
)

// Item words, one home each.
const (
	labelShow        = "Show ribbon"
	labelHide        = "Hide ribbon"
	labelAddClock    = "Add clock"
	labelSettings    = "Settings"
	labelPosition    = "Position"
	labelLeftEdge    = "Centre on left edge"
	labelRightEdge   = "Centre on right edge"
	labelTopEdge     = "Centre on top edge"
	labelBottomEdge  = "Centre on bottom edge"
	labelAlwaysOnTop = "Always on top"
	labelPin         = "Pin ribbon"
	labelSunMap      = "Sun map"
	labelHelp        = "Help"
	labelAbout       = "About"
	labelLicence     = "Licence"
	labelUpdates     = "Check for updates"
	labelExit        = "Exit"
)

// TrayMenu answers the tray menu for a ribbon that is or is not visible (FR-502): the visibility
// item names the opposite of what is, so it says what pressing it will do.
func (s *Service) TrayMenu(visible bool) []menus.Item {
	toggle := menus.Item{Action: menus.Show, Label: labelShow}
	if visible {
		toggle = menus.Item{Action: menus.Hide, Label: labelHide}
	}
	return []menus.Item{
		toggle, addClockItem(), settingsItem(), s.styleItem(), s.colourItem(), s.orientationItem(), s.positionItem(),
		s.alwaysOnTopItem(), s.pinItem(), s.sunMapItem(), helpItem(), exitItem(),
	}
}

// ContextMenu answers the menu the ribbon offers when right-clicked (FR-108).
func (s *Service) ContextMenu() []menus.Item {
	return []menus.Item{
		addClockItem(), settingsItem(), s.styleItem(), s.colourItem(), s.orientationItem(), s.positionItem(),
		s.alwaysOnTopItem(), s.pinItem(), s.sunMapItem(), helpItem(), {Action: menus.Hide, Label: labelHide}, exitItem(),
	}
}

// SettingsChoices answers every choice the menus offer, for Settings to offer as well (FR-624): the
// same items both menus hold, so their words and ticks have one home. What is left of the menus is
// commands (show or hide, Add clock, Settings, Help, Exit), which choose nothing.
func (s *Service) SettingsChoices() []menus.Item {
	return []menus.Item{
		s.styleItem(), s.colourItem(), s.orientationItem(), s.positionItem(),
		s.alwaysOnTopItem(), s.pinItem(), s.sunMapItem(),
	}
}

// positionItem is the Position submenu both menus hold (FR-408): the two edges along which the
// ribbon runs its length, so a vertical ribbon is offered the left and right edges and a horizontal
// one the top and bottom.
func (s *Service) positionItem() menus.Item {
	children := []menus.Item{{Action: menus.TopEdge, Label: labelTopEdge}, {Action: menus.BottomEdge, Label: labelBottomEdge}}
	if s.Settings().Orientation == ribbon.Vertical {
		children = []menus.Item{{Action: menus.LeftEdge, Label: labelLeftEdge}, {Action: menus.RightEdge, Label: labelRightEdge}}
	}
	return menus.Item{Label: labelPosition, Children: children}
}

// exitItem ends the application, from either menu (FR-108, FR-502).
func exitItem() menus.Item { return menus.Item{Action: menus.Exit, Label: labelExit} }

// helpItem is the Help submenu both menus hold (FR-508, FR-509).
func helpItem() menus.Item {
	return menus.Item{Label: labelHelp, Children: []menus.Item{
		{Action: menus.About, Label: labelAbout},
		{Action: menus.Licence, Label: labelLicence},
		{Action: menus.Updates, Label: labelUpdates},
	}}
}

// CloseRequested answers what a request to close the ribbon does, such as Alt+F4: it hides the
// ribbon and the application keeps running; only Exit ends it (FR-504, FR-507).
func (s *Service) CloseRequested() menus.Action {
	return menus.Hide
}

func addClockItem() menus.Item { return menus.Item{Action: ActionAddClock, Label: labelAddClock} }

func settingsItem() menus.Item { return menus.Item{Action: menus.Settings, Label: labelSettings} }

func (s *Service) alwaysOnTopItem() menus.Item {
	return menus.Item{
		Action: menus.AlwaysOnTop, Label: labelAlwaysOnTop,
		Checkable: true, Checked: s.Settings().AlwaysOnTop,
	}
}

// pinItem follows Always on top in both menus, ticked while the ribbon is pinned (FR-613).
func (s *Service) pinItem() menus.Item {
	return menus.Item{Action: menus.Pin, Label: labelPin, Checkable: true, Checked: s.Settings().Pinned}
}
