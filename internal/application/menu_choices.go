package application

import (
	"slices"
	"strings"

	"github.com/oernster/timeribbon/internal/domain/settings"
)

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

// colourPrefix begins each Colour item's action, which ends in the scheme it chooses.
const colourPrefix = "colour-"

// colourLabels are the Colour items' words, one home each.
var colourLabels = map[settings.Colour]string{
	settings.Classic: "Classic", settings.Neon: "Neon", settings.Ocean: "Ocean",
	settings.Sunset: "Sunset", settings.Forest: "Forest",
}

// labelColour is the Colour submenu's own word.
const labelColour = "Colour"

// ColourOf answers the scheme a Colour item chooses; false for any other action.
func ColourOf(action MenuAction) (settings.Colour, bool) {
	name, found := strings.CutPrefix(string(action), colourPrefix)
	colour := settings.Colour(name)
	return colour, found && slices.Contains(settings.Colours, colour)
}

// colourItem is the Colour submenu both menus hold, the current scheme ticked (FR-611).
func (s *Service) colourItem() MenuItem {
	current := s.Settings().Colour
	children := make([]MenuItem, 0, len(settings.Colours))
	for _, colour := range settings.Colours {
		children = append(children, MenuItem{
			Action: MenuAction(colourPrefix + string(colour)), Label: colourLabels[colour],
			Checkable: true, Checked: colour == current,
		})
	}
	return MenuItem{Label: labelColour, Children: children}
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
