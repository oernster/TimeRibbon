package main

import (
	"fmt"
	"slices"

	"github.com/oernster/timeribbon/internal/application"
	"github.com/oernster/timeribbon/ribbonkit/application/menus"
)

// Choose carries out one of the menus' choices from Settings (FR-624), exactly as the menu item
// would; an action that is not one of the choices Settings offers now is refused.
func (a *App) Choose(action string) error {
	chosen := menus.Action(action)
	if !slices.Contains(choiceActions(a.service.SettingsChoices()), chosen) {
		return fmt.Errorf("%w: a choice named %q", application.ErrUnknownChoice, action)
	}
	a.act(chosen)
	return nil
}

// choiceActions answers the action of every item in items that has one, groups' children included.
func choiceActions(items []menus.Item) []menus.Action {
	var actions []menus.Action
	for _, item := range items {
		if item.Action != "" {
			actions = append(actions, item.Action)
		}
		actions = append(actions, choiceActions(item.Children)...)
	}
	return actions
}
