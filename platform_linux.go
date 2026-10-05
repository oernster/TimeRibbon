package main

import "github.com/oernster/ribbonkit/infrastructure/gtkmain"

// On Linux GTK is sent through X11 before Wails opens it: the ribbon must choose where it stands,
// which a window on Wayland may not. The web view's DMABUF renderer is turned off before it starts,
// since on NVIDIA's own driver it draws nothing.
func init() {
	gtkmain.ForceX11()
	gtkmain.AvoidDMABUF()
}
