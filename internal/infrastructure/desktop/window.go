// Package desktop is TimeRibbon's integration with the desktop below the window (CON-7): the
// notification area icon and its menu (FR-501 to FR-503), the end of a move (FR-404), the display,
// time and resume broadcasts (FR-209, FR-406) and the operations on the ribbon's own window (FR-101,
// FR-401, FR-405). Each platform supplies them in files of its own; this one holds what they share.
package desktop

// Window is the ribbon's window as the platform names it: a window handle on Windows. Zero is no
// window, which is what the facade holds until startup has found it.
type Window uintptr
