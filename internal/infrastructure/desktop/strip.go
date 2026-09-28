package desktop

import (
	"errors"
	"fmt"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/oernster/timeribbon/internal/domain/placement"
)

// findAttempts and findPause bound the wait for the strip's window to exist after Wails starts.
const (
	findAttempts = 50
	findPause    = 20 * time.Millisecond
)

// ErrStripNotFound is answered when no window of the strip's class appears.
var ErrStripNotFound = errors.New("the strip's window was not found")

// FindStrip answers the window of class, the class name the strip's window is created with.
func FindStrip(class string) (windows.HWND, error) {
	for range findAttempts {
		if handle, _, _ := procFindWindow.Call(utf16Pointer(class), 0); handle != 0 {
			return windows.HWND(handle), nil
		}
		time.Sleep(findPause)
	}
	return 0, fmt.Errorf("%w: class %s", ErrStripNotFound, class)
}

// HideFromTaskbar makes the strip a tool window, which has no taskbar button (FR-101). Wails creates
// its window with WS_EX_APPWINDOW, which forces a button, so that flag is taken off. It is done
// while the window is hidden, since the taskbar reads the style when a window is shown.
func HideFromTaskbar(strip windows.HWND) error {
	style, _, _ := procGetWindowLongPtr.Call(uintptr(strip), uintptr(exStyleIndex))
	// SetWindowLongPtr answers the previous style, never zero here since Wails sets WS_EX_APPWINDOW;
	// zero is the failure.
	if previous, _, err := procSetWindowLongPtr.Call(uintptr(strip), uintptr(exStyleIndex), style&^wsExAppWindow|wsExToolWindow); previous == 0 {
		return fmt.Errorf("changing the strip's window style: %w", err)
	}
	return nil
}

// exStyleIndex is GWL_EXSTYLE held in a variable, since a negative constant cannot convert to the
// unsigned argument the call takes.
var exStyleIndex int32 = gwlExStyle

// Place moves and sizes the strip in physical pixels without raising or activating it (FR-405).
func Place(strip windows.HWND, at placement.Point, size placement.Size) error {
	ok, _, err := procSetWindowPos.Call(uintptr(strip), 0,
		uintptr(at.X), uintptr(at.Y), uintptr(size.Width), uintptr(size.Height),
		swpNoZOrder|swpNoActivate|swpFrameChanged)
	if ok == 0 {
		return fmt.Errorf("placing the strip: %w", err)
	}
	return nil
}

// Position answers the strip's top-left corner in physical pixels.
func Position(strip windows.HWND) (placement.Point, error) {
	var bounds rect
	if ok, _, err := procGetWindowRect.Call(uintptr(strip), uintptr(unsafe.Pointer(&bounds))); ok == 0 {
		return placement.Point{}, fmt.Errorf("reading the strip's position: %w", err)
	}
	return placement.Point{X: int(bounds.left), Y: int(bounds.top)}, nil
}

// DragThreshold answers how far the pointer must move, in DIP, before a press becomes a drag: the
// Windows drag rectangle at 100 percent scaling (FR-401).
func DragThreshold() placement.Size {
	width, _, _ := procGetSystemMetricsForDpi.Call(smCxDrag, placement.BaseDPI)
	height, _, _ := procGetSystemMetricsForDpi.Call(smCyDrag, placement.BaseDPI)
	return placement.Size{Width: int(width), Height: int(height)}
}
