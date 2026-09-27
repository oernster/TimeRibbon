package appdata

import (
	"errors"
	"testing"
)

func TestTheFolderIsTimeStripUnderAppData(t *testing.T) {
	t.Parallel()
	got, err := Dir(func(string) (string, bool) { return `C:\Users\Someone\AppData\Roaming`, true })
	if err != nil || got != `C:\Users\Someone\AppData\Roaming\TimeStrip` {
		t.Errorf("got %q (%v)", got, err)
	}
}

func TestAnUnsetOrEmptyVariableIsRefused(t *testing.T) {
	t.Parallel()
	for _, lookup := range []func(string) (string, bool){
		func(string) (string, bool) { return "", false },
		func(string) (string, bool) { return "", true },
	} {
		if _, err := Dir(lookup); !errors.Is(err, ErrNoAppData) {
			t.Errorf("got %v", err)
		}
	}
}
