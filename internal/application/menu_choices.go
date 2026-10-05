package application

import (
	"github.com/oernster/ribbonkit/application/menus"
	"github.com/oernster/timeribbon/internal/domain/settings"
)

// The actions of the Style submenu (FR-108, FR-502), TimeRibbon's own; the Colour and Orientation
// submenus are ribbonkit's.
const (
	ActionDigital  menus.Action = "digital"
	ActionAnalogue menus.Action = "analogue"
)

// Their words, one home each.
const (
	labelStyle    = "Style"
	labelDigital  = "Digital"
	labelAnalogue = "Analogue"
)

// styleActions maps each Style item to the style it chooses.
var styleActions = map[menus.Action]settings.Style{ActionDigital: settings.Digital, ActionAnalogue: settings.Analogue}

// StyleOf answers the style a Style item chooses; false for any other action.
func StyleOf(action menus.Action) (settings.Style, bool) {
	style, ok := styleActions[action]
	return style, ok
}

// styleItem is the Style submenu both menus hold, the current style ticked.
func (s *Service) styleItem() menus.Item {
	current := s.Settings().Style
	return menus.Item{Label: labelStyle, Children: []menus.Item{
		{Action: ActionDigital, Label: labelDigital, Checkable: true, Checked: current == settings.Digital},
		{Action: ActionAnalogue, Label: labelAnalogue, Checkable: true, Checked: current == settings.Analogue},
	}}
}
