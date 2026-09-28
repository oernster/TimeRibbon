package appdata

import (
	"errors"
	"testing"
)

func TestTheFolderIsTimeRibbonUnderApplicationSupport(t *testing.T) {
	t.Parallel()
	lookup := func(name string) (string, bool) {
		if name == homeVariable {
			return "/Users/someone", true
		}
		return "", false
	}
	want := "/Users/someone/Library/Application Support/TimeRibbon"
	if got, err := Dir(lookup); err != nil || got != want {
		t.Errorf("got %q (%v), want %q", got, err, want)
	}
}

func TestWithNoHomeThereIsNoFolder(t *testing.T) {
	t.Parallel()
	if _, err := Dir(func(string) (string, bool) { return "", false }); !errors.Is(err, ErrNoAppData) {
		t.Errorf("got %v", err)
	}
}
