package application

import "github.com/oernster/timeribbon/internal/domain/settings"

// The actions of the Style and Orientation submenus (FR-108, FR-502).
const (
	ActionDigital    MenuAction = "digital"
	ActionAnalogue   MenuAction = "analogue"
	ActionHorizontal MenuAction = "horizontal"
	ActionVertical   MenuAction = "vertical"
)

// Their words, one home each.
const (
	labelStyle       = "Style"
	labelDigital     = "Digital"
	labelAnalogue    = "Analogue"
	labelOrientation = "Orientation"
	labelHorizontal  = "Horizontal"
	labelVertical    = "Vertical"
)

// styleActions maps each Style item to the style it chooses.
var styleActions = map[MenuAction]settings.Style{ActionDigital: settings.Digital, ActionAnalogue: settings.Analogue}

// orientationActions maps each Orientation item to the orientation it chooses.
var orientationActions = map[MenuAction]settings.Orientation{
	ActionHorizontal: settings.Horizontal, ActionVertical: settings.Vertical,
}

// StyleOf answers the style a Style item chooses; false for any other action.
func StyleOf(action MenuAction) (settings.Style, bool) {
	style, ok := styleActions[action]
	return style, ok
}

// OrientationOf answers the orientation an Orientation item chooses; false for any other action.
func OrientationOf(action MenuAction) (settings.Orientation, bool) {
	orientation, ok := orientationActions[action]
	return orientation, ok
}

// styleItem is the Style submenu both menus hold, the current style ticked.
func (s *Service) styleItem() MenuItem {
	current := s.Settings().Style
	return MenuItem{Label: labelStyle, Children: []MenuItem{
		{Action: ActionDigital, Label: labelDigital, Checkable: true, Checked: current == settings.Digital},
		{Action: ActionAnalogue, Label: labelAnalogue, Checkable: true, Checked: current == settings.Analogue},
	}}
}

// orientationItem is the Orientation submenu both menus hold, the current orientation ticked.
func (s *Service) orientationItem() MenuItem {
	current := s.Settings().Orientation
	return MenuItem{Label: labelOrientation, Children: []MenuItem{
		{Action: ActionHorizontal, Label: labelHorizontal, Checkable: true, Checked: current == settings.Horizontal},
		{Action: ActionVertical, Label: labelVertical, Checkable: true, Checked: current == settings.Vertical},
	}}
}
