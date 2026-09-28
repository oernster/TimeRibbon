package desktop

import "github.com/oernster/timeribbon/internal/application"

// numbered answers the action of every item a native menu gives a number, in the order they are
// given out: depth first, so a submenu's items follow every item before it. A submenu takes no
// number of its own (FR-508).
func numbered(items []application.MenuItem) []application.MenuAction {
	var out []application.MenuAction
	for _, item := range items {
		if len(item.Children) > 0 {
			out = append(out, numbered(item.Children)...)
			continue
		}
		out = append(out, item.Action)
	}
	return out
}

// actionAt answers the action numbered index; false for a number no item carries.
func actionAt(items []application.MenuItem, index int) (application.MenuAction, bool) {
	actions := numbered(items)
	if index < 0 || index >= len(actions) {
		return "", false
	}
	return actions[index], true
}
