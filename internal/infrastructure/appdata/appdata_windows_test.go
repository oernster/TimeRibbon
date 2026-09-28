package appdata

import "testing"

func TestTheFolderIsTimeRibbonUnderAppData(t *testing.T) {
	t.Parallel()
	got, err := Dir(func(string) (string, bool) { return `C:\Users\Someone\AppData\Roaming`, true })
	if err != nil || got != `C:\Users\Someone\AppData\Roaming\TimeRibbon` {
		t.Errorf("got %q (%v)", got, err)
	}
}
