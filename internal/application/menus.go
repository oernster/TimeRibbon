package application

import (
	"github.com/oernster/timestrip/internal/domain/placement"
	"github.com/oernster/timestrip/internal/domain/settings"
)

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
	ActionAbout       MenuAction = "about"
	ActionLicence     MenuAction = "licence"
	ActionExit        MenuAction = "exit"
	ActionLeftEdge    MenuAction = "left-edge"
	ActionRightEdge   MenuAction = "right-edge"
	ActionTopEdge     MenuAction = "top-edge"
	ActionBottomEdge  MenuAction = "bottom-edge"
)

// edgeActions maps each Position item to the edge it puts the strip against (FR-408).
var edgeActions = map[MenuAction]placement.Edge{
	ActionLeftEdge:   placement.Left,
	ActionRightEdge:  placement.Right,
	ActionTopEdge:    placement.Top,
	ActionBottomEdge: placement.Bottom,
}

// EdgeOf answers the edge a Position item puts the strip against; false for any other action.
func EdgeOf(action MenuAction) (placement.Edge, bool) {
	edge, ok := edgeActions[action]
	return edge, ok
}

// MenuItem is one entry of a menu.
type MenuItem struct {
	Action MenuAction
	Label  string
	// Checkable items show Checked beside their label.
	Checkable bool
	Checked   bool
	// Children makes the item a submenu holding them; such an item has no action of its own.
	Children []MenuItem
}

// Item words, one home each.
const (
	labelShow        = "Show strip"
	labelHide        = "Hide strip"
	labelAddClock    = "Add clock"
	labelSettings    = "Settings"
	labelPosition    = "Position"
	labelLeftEdge    = "Centre on left edge"
	labelRightEdge   = "Centre on right edge"
	labelTopEdge     = "Centre on top edge"
	labelBottomEdge  = "Centre on bottom edge"
	labelAlwaysOnTop = "Always on top"
	labelHelp        = "Help"
	labelAbout       = "About"
	labelLicence     = "Licence"
	labelExit        = "Exit"
)

// TrayMenu answers the tray menu for a strip that is or is not visible (FR-502): the visibility
// item names the opposite of what is, so it says what pressing it will do.
func (s *Service) TrayMenu(visible bool) []MenuItem {
	toggle := MenuItem{Action: ActionShow, Label: labelShow}
	if visible {
		toggle = MenuItem{Action: ActionHide, Label: labelHide}
	}
	return []MenuItem{toggle, addClockItem(), settingsItem(), s.positionItem(), s.alwaysOnTopItem(), helpItem(), exitItem()}
}

// ContextMenu answers the menu the strip offers when right-clicked (FR-108).
func (s *Service) ContextMenu() []MenuItem {
	return []MenuItem{
		addClockItem(), settingsItem(), s.positionItem(), s.alwaysOnTopItem(), helpItem(),
		{Action: ActionHide, Label: labelHide}, exitItem(),
	}
}

// positionItem is the Position submenu both menus hold (FR-408): the two edges along which the
// strip runs its length, so a vertical strip is offered the left and right edges and a horizontal
// one the top and bottom.
func (s *Service) positionItem() MenuItem {
	children := []MenuItem{{Action: ActionTopEdge, Label: labelTopEdge}, {Action: ActionBottomEdge, Label: labelBottomEdge}}
	if s.Settings().Orientation == settings.Vertical {
		children = []MenuItem{{Action: ActionLeftEdge, Label: labelLeftEdge}, {Action: ActionRightEdge, Label: labelRightEdge}}
	}
	return MenuItem{Label: labelPosition, Children: children}
}

// exitItem ends the application, from either menu (FR-108, FR-502).
func exitItem() MenuItem { return MenuItem{Action: ActionExit, Label: labelExit} }

// helpItem is the Help submenu both menus hold (FR-508).
func helpItem() MenuItem {
	return MenuItem{Label: labelHelp, Children: []MenuItem{
		{Action: ActionAbout, Label: labelAbout},
		{Action: ActionLicence, Label: labelLicence},
	}}
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
