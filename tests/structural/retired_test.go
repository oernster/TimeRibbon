package structural

// The product carried another name until 2.0.0 and was renamed over a trademark concern. Nothing
// carries over from the old name (Oliver, 2026-09-28), so no file has a reason to hold it: this test
// fails on any tracked text file that does, so a rename that misses a corner is caught here rather
// than shipped. The name is only ever spelt in pieces, so this file does not hold it either. Git
// lists only tracked files, so a new file is checked once it is added.

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// retiredNames are the names no tracked file may hold, compared without regard to case.
var retiredNames = []string{"time" + "strip"}

// trackedTextFiles answers every file git tracks that holds text, as paths from the repository's
// root. A file holding a NUL byte is binary and is left out.
func trackedTextFiles(t *testing.T) []string {
	t.Helper()
	root := repoRoot(t)
	command := exec.Command("git", "ls-files", "-z")
	command.Dir = root
	listing, err := command.Output()
	if err != nil {
		t.Fatalf("listing the tracked files: %v", err)
	}
	var out []string
	for _, name := range strings.Split(string(listing), "\x00") {
		if name == "" {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatalf("reading %s: %v", name, err)
		}
		if !bytes.Contains(raw, []byte{0}) {
			out = append(out, name)
		}
	}
	return out
}

// holdsRetiredName answers the first retired name text holds; empty when it holds none.
func holdsRetiredName(text string) string {
	lowered := strings.ToLower(text)
	for _, name := range retiredNames {
		if strings.Contains(lowered, name) {
			return name
		}
	}
	return ""
}

func TestNoTrackedFileHoldsARetiredName(t *testing.T) {
	root := repoRoot(t)
	files := trackedTextFiles(t)
	if len(files) == 0 {
		t.Fatal("git listed no tracked text files, so nothing was checked")
	}
	for _, name := range files {
		raw, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatalf("reading %s: %v", name, err)
		}
		for number, line := range strings.Split(string(raw), "\n") {
			if found := holdsRetiredName(line); found != "" {
				t.Errorf("%s:%d holds the retired name %q", name, number+1, found)
			}
		}
	}
}

func TestRetiredNamesAreFoundInAnyCase(t *testing.T) {
	for _, text := range []string{"Time" + "Strip", "time" + "strip.exe", "TIME" + "STRIP_RUNLOG"} {
		if holdsRetiredName(text) == "" {
			t.Errorf("%q was not recognised", text)
		}
	}
	if found := holdsRetiredName("TimeRibbon shows a strip of clocks"); found != "" {
		t.Errorf("the current name was taken for %q", found)
	}
}
