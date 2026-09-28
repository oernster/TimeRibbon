package desktop

import (
	"fmt"
	"os/exec"
)

// opener is the freedesktop program that opens an address with whatever the desktop has chosen for
// it. Inside a Flatpak it reaches the host through the OpenURI portal.
const opener = "xdg-open"

// OpenInBrowser hands address to the program the desktop opens such addresses with, answering an
// error when it cannot: no opener installed, nothing to open the address with. The application
// never fetches the address itself.
func OpenInBrowser(address string) error { return openWith(opener, address) }

// openWith runs program on address, naming the address in any refusal.
func openWith(program, address string) error {
	if err := exec.Command(program, address).Run(); err != nil {
		return fmt.Errorf("opening %s: %w", address, err)
	}
	return nil
}
