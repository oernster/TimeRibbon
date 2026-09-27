package clock

import "strings"

// MaxLabelLength is the most characters a label holds (FR-307).
const MaxLabelLength = 32

// DefaultLabel answers the label a zone is shown by until the user types another: the last
// segment of its id with underscores read as spaces. "America/Argentina/Buenos_Aires" gives
// "Buenos Aires".
func DefaultLabel(zone string) string {
	last := zone[strings.LastIndex(zone, "/")+1:]
	return strings.ReplaceAll(last, "_", " ")
}

// Label answers the label to store for what the user typed (FR-303, FR-307): trimmed of
// surrounding spaces, cut to MaxLabelLength characters; the zone's default label when nothing
// is left.
func Label(typed, zone string) string {
	trimmed := strings.TrimSpace(typed)
	if trimmed == "" {
		return DefaultLabel(zone)
	}
	characters := []rune(trimmed)
	if len(characters) > MaxLabelLength {
		return strings.TrimSpace(string(characters[:MaxLabelLength]))
	}
	return trimmed
}
