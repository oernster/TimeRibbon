package product

import (
	"strings"
	"testing"
)

// The address is asserted literally, so a typo fails here rather than sending a supporter to a page
// that is not the author's.
func TestTheDonateAddressIsTheAuthorsAndSecure(t *testing.T) {
	t.Parallel()
	if DonateURL != "https://www.paypal.com/ncp/payment/THUS4KZ5GECH8" {
		t.Errorf("DonateURL is %q", DonateURL)
	}
	if !strings.HasPrefix(DonateURL, "https://") {
		t.Errorf("DonateURL is not https: %q", DonateURL)
	}
}
