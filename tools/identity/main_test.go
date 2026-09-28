package main

import (
	"strings"
	"testing"
)

// The lock's name follows Wails' formula, as its source builds it. Measured 2026-09-28: the
// installed Flatpak owned exactly this name.
func TestTheSingleInstanceNameFollowsWails(t *testing.T) {
	t.Parallel()
	if got := singleInstanceName("uk.codecrafter.TimeRibbon"); got != "org.wails_app_uk_codecrafter_TimeRibbon.SingleInstance" {
		t.Errorf("got %q", got)
	}
}

func TestEveryNameIsPrintedForTheShell(t *testing.T) {
	t.Parallel()
	var out strings.Builder
	print(&out)
	for _, want := range []string{"APP_NAME='TimeRibbon'\n", "APP_ID='uk.codecrafter.TimeRibbon'\n", "BIN_NAME='timeribbon'\n", "SINGLE_INSTANCE_NAME='org.wails_app_uk_codecrafter_TimeRibbon.SingleInstance'\n"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("missing %q in %q", want, out.String())
		}
	}
}
