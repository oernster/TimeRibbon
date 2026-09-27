package application

// MenuAction names what a menu item does. The window and the tray act on it; the words shown for
// it live here alone.
type MenuAction string

// The actions the tray and context menus offer.
const (
	ActionShow        MenuAction = "show"
	ActionHide        MenuAction = "hide"
	ActionAddClock    MenuAction = "add-clock"
	ActionSettings    MenuAction = "settings"
	ActionAlwaysOnTop MenuAction = "always-on-top"
	ActionExit        MenuAction = "exit"
)

// MenuItem is one entry of a menu.
type MenuItem struct {
	Action MenuAction
	Label  string
	// Checkable items show Checked beside their label.
	Checkable bool
	Checked   bool
}

// Item words, one home each.
const (
	labelShow        = "Show strip"
	labelHide        = "Hide strip"
	labelAddClock    = "Add clock"
	labelSettings    = "Settings"
	labelAlwaysOnTop = "Always on top"
	labelExit        = "Exit"
)

// TrayMenu answers the tray menu for a strip that is or is not visible (FR-502): the visibility
// item names the opposite of what is, so it says what pressing it will do.
func (s *Service) TrayMenu(visible bool) []MenuItem {
	toggle := MenuItem{Action: ActionShow, Label: labelShow}
	if visible {
		toggle = MenuItem{Action: ActionHide, Label: labelHide}
	}
	return []MenuItem{toggle, addClockItem(), settingsItem(), s.alwaysOnTopItem(), {Action: ActionExit, Label: labelExit}}
}

// ContextMenu answers the menu the strip offers when right-clicked (FR-108).
func (s *Service) ContextMenu() []MenuItem {
	return []MenuItem{addClockItem(), settingsItem(), s.alwaysOnTopItem(), {Action: ActionHide, Label: labelHide}}
}

// CloseRequested answers what a request to close the strip does, such as Alt+F4: it hides the
// strip and the application keeps running; only Exit ends it (FR-504, FR-507).
func (s *Service) CloseRequested() MenuAction {
	return ActionHide
}

func addClockItem() MenuItem { return MenuItem{Action: ActionAddClock, Label: labelAddClock} }

func settingsItem() MenuItem { return MenuItem{Action: ActionSettings, Label: labelSettings} }

func (s *Service) alwaysOnTopItem() MenuItem {
	return MenuItem{
		Action: ActionAlwaysOnTop, Label: labelAlwaysOnTop,
		Checkable: true, Checked: s.Settings().AlwaysOnTop,
	}
}
