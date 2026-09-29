package main

import (
	"fmt"
	"slices"

	"github.com/oernster/timeribbon/internal/application"
)

// Choose carries out one of the menus' choices from Settings (FR-624), exactly as the menu item
// would; an action that is not one of the choices Settings offers now is refused.
func (a *App) Choose(action string) error {
	chosen := application.MenuAction(action)
	if !slices.Contains(choiceActions(a.service.SettingsChoices()), chosen) {
		return fmt.Errorf("%w: a choice named %q", application.ErrUnknownChoice, action)
	}
	a.act(chosen)
	return nil
}

// choiceActions answers the action of every item in items that has one, groups' children included.
func choiceActions(items []application.MenuItem) []application.MenuAction {
	var actions []application.MenuAction
	for _, item := range items {
		if item.Action != "" {
			actions = append(actions, item.Action)
		}
		actions = append(actions, choiceActions(item.Children)...)
	}
	return actions
}
