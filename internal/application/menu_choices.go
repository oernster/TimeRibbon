package application

import (
	"slices"
	"strings"

	"github.com/oernster/timeribbon/internal/domain/settings"
	"github.com/oernster/timeribbon/ribbonkit/application/menus"
	"github.com/oernster/timeribbon/ribbonkit/domain/ribbon"
)

// The actions of the Style and Orientation submenus (FR-108, FR-502).
const (
	ActionDigital    menus.Action = "digital"
	ActionAnalogue   menus.Action = "analogue"
	ActionHorizontal menus.Action = "horizontal"
	ActionVertical   menus.Action = "vertical"
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
var styleActions = map[menus.Action]settings.Style{ActionDigital: settings.Digital, ActionAnalogue: settings.Analogue}

// orientationActions maps each Orientation item to the orientation it chooses.
var orientationActions = map[menus.Action]ribbon.Orientation{
	ActionHorizontal: ribbon.Horizontal, ActionVertical: ribbon.Vertical,
}

// colourPrefix begins each Colour item's action, which ends in the scheme it chooses.
const colourPrefix = "colour-"

// colourLabels are the Colour items' words, one home each.
var colourLabels = map[ribbon.Colour]string{
	ribbon.Classic: "Classic", ribbon.Neon: "Neon", ribbon.Ocean: "Ocean",
	ribbon.Sunset: "Sunset", ribbon.Forest: "Forest", ribbon.Amber: "Amber",
	ribbon.Ruby: "Ruby", ribbon.Indigo: "Indigo", ribbon.Berry: "Berry",
	ribbon.Contrast: "Contrast",
}

// labelColour is the Colour submenu's own word.
const labelColour = "Colour"

// ColourOf answers the scheme a Colour item chooses; false for any other action.
func ColourOf(action menus.Action) (ribbon.Colour, bool) {
	name, found := strings.CutPrefix(string(action), colourPrefix)
	colour := ribbon.Colour(name)
	return colour, found && slices.Contains(ribbon.Colours, colour)
}

// colourItem is the Colour submenu both menus hold, the current scheme ticked (FR-611).
func (s *Service) colourItem() menus.Item {
	current := s.Settings().Colour
	children := make([]menus.Item, 0, len(ribbon.Colours))
	for _, colour := range ribbon.Colours {
		children = append(children, menus.Item{
			Action: menus.Action(colourPrefix + string(colour)), Label: colourLabels[colour],
			Checkable: true, Checked: colour == current,
		})
	}
	return menus.Item{Label: labelColour, Children: children}
}

// StyleOf answers the style a Style item chooses; false for any other action.
func StyleOf(action menus.Action) (settings.Style, bool) {
	style, ok := styleActions[action]
	return style, ok
}

// OrientationOf answers the orientation an Orientation item chooses; false for any other action.
func OrientationOf(action menus.Action) (ribbon.Orientation, bool) {
	orientation, ok := orientationActions[action]
	return orientation, ok
}

// styleItem is the Style submenu both menus hold, the current style ticked.
func (s *Service) styleItem() menus.Item {
	current := s.Settings().Style
	return menus.Item{Label: labelStyle, Children: []menus.Item{
		{Action: ActionDigital, Label: labelDigital, Checkable: true, Checked: current == settings.Digital},
		{Action: ActionAnalogue, Label: labelAnalogue, Checkable: true, Checked: current == settings.Analogue},
	}}
}

// orientationItem is the Orientation submenu both menus hold, the current orientation ticked.
func (s *Service) orientationItem() menus.Item {
	current := s.Settings().Orientation
	return menus.Item{Label: labelOrientation, Children: []menus.Item{
		{Action: ActionHorizontal, Label: labelHorizontal, Checkable: true, Checked: current == ribbon.Horizontal},
		{Action: ActionVertical, Label: labelVertical, Checkable: true, Checked: current == ribbon.Vertical},
	}}
}
