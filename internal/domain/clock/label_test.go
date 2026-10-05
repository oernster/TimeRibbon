package clock

import "testing"

func TestDefaultLabelIsTheLastSegmentWithSpaces(t *testing.T) {
	t.Parallel()
	for zone, want := range map[string]string{
		"America/Argentina/Buenos_Aires": "Buenos Aires",
		"America/New_York":               "New York",
		"UTC":                            "UTC",
	} {
		if got := DefaultLabel(zone); got != want {
			t.Errorf("DefaultLabel(%q) = %q, want %q", zone, got, want)
		}
	}
}

// FR-303.
func TestEmptyLabelFallsBackToDefault(t *testing.T) {
	t.Parallel()
	for _, typed := range []string{"", "   ", "\t"} {
		if got := Label(typed, "Europe/London"); got != "London" {
			t.Errorf("Label(%q) = %q, want London", typed, got)
		}
	}
	if got := Label("  Brighton ", "Europe/London"); got != "Brighton" {
		t.Errorf("a typed label is kept trimmed: got %q", got)
	}
}
