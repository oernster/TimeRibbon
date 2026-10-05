package product

import (
	"strings"
	"testing"
)

// The kit is handed the product's own name and application id, the ones every other surface uses.
func TestTheKitIsHandedTheProductsNames(t *testing.T) {
	t.Parallel()
	if app := App(); app.Name != Name || app.AppID != AppID {
		t.Errorf("App answered %+v", app)
	}
}

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
