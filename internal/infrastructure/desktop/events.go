package desktop

import "github.com/oernster/timeribbon/internal/application"

// eventBuffer is how many events may wait unread before the next is dropped rather than block the
// thread the desktop calls in on.
const eventBuffer = 32

// EventKind names what happened.
type EventKind int

// The events the desktop reports.
const (
	// EventMenu is a menu item chosen; Event.Action names it.
	EventMenu EventKind = iota
	// EventIconClicked is a left click on the tray icon (FR-503).
	EventIconClicked
	// EventMoveEnded is the end of a move of the ribbon's window (FR-404).
	EventMoveEnded
	// EventDisplayChanged is a change of displays, resolution or arrangement (FR-406).
	EventDisplayChanged
	// EventTimeChanged is a change of the system time or time zone (FR-209).
	EventTimeChanged
	// EventResumed is a resume from sleep (FR-209).
	EventResumed
	// EventPointerArrived is the pointer come onto the ribbon or its tab (FR-615).
	EventPointerArrived
	// EventPointerLeft is the pointer gone off the ribbon or its tab (FR-616).
	EventPointerLeft
	// EventMenuClosed is a popup menu closed, chosen from or not; one open holds the ribbon (FR-616).
	EventMenuClosed
)

// Event is one thing that happened on the desktop.
type Event struct {
	Kind   EventKind
	Action application.MenuAction
}
