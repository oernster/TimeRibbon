package clock

import (
	"strings"

	"github.com/oernster/ribbonkit/domain/ribbon"
)

// DefaultLabel answers the label a zone is shown by until the user types another: the last
// segment of its id with underscores read as spaces. "America/Argentina/Buenos_Aires" gives
// "Buenos Aires".
func DefaultLabel(zone string) string {
	last := zone[strings.LastIndex(zone, "/")+1:]
	return strings.ReplaceAll(last, "_", " ")
}

// Label answers the label to store for what the user typed (FR-303, FR-307): ribbonkit's rule for
// every cell's label, trimmed and capped, with the zone's default label when nothing is left.
func Label(typed, zone string) string {
	return ribbon.Label(typed, DefaultLabel(zone))
}
