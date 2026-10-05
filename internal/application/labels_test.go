package application

import (
	"github.com/oernster/ribbonkit/application/menus"
	"github.com/oernster/ribbonkit/domain/ribbon"
)

// The words of every ribbon's items, which live in ribbonkit's menus package, read from its items so
// these tests name them as they did when they were TimeRibbon's.
var (
	labelShow        = menus.Visibility(false).Label
	labelHide        = menus.HideItem().Label
	labelSettings    = menus.SettingsItem().Label
	labelHelp        = menus.HelpItem().Label
	labelExit        = menus.ExitItem().Label
	labelAlwaysOnTop = menus.AlwaysOnTopItem(false).Label
	labelPin         = menus.PinItem(false).Label
	labelColour      = menus.ColourItem(ribbon.Classic).Label
	labelOrientation = menus.OrientationItem(ribbon.Vertical).Label
	colourLabels     = colourWords()
)

// colourWords answers each scheme's word as the Colour submenu shows it.
func colourWords() map[ribbon.Colour]string {
	words := map[ribbon.Colour]string{}
	for _, child := range menus.ColourItem(ribbon.Classic).Children {
		colour, _ := menus.ColourOf(child.Action)
		words[colour] = child.Label
	}
	return words
}
