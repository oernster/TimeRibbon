package structural

// The product carried another name until 2.0.0 and was renamed over a trademark concern. Its window
// was called by the word that name ended in; it is now the ribbon. Nothing carries over from the
// old names (Oliver, 2026-09-28), so no file has a reason to hold them: this test fails on any
// tracked text file that does, so a rename that misses a corner is caught here rather than shipped.
// The retired word is only ever spelt in pieces, so this file does not hold it either. New files
// are checked before they are added, so the check holds on a working tree as well as on a commit.

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/oernster/ribbonkit/structure"
)

// retiredWord is the word no tracked file may hold, compared without regard to case. The old
// product name contains it, so it retires both.
var retiredWord = "st" + "rip"

// thirdPartyFiles are tracked files other people's names fill: the lock file names npm packages,
// one of which has the retired word in its own name.
var thirdPartyFiles = []string{"frontend/package-lock.json"}

// trackedTextFiles answers every file git tracks or would track (a new file not yet added, unless
// it is ignored) that holds text, as paths from the repository's root. A file holding a NUL byte is
// binary and is left out; one git still lists that is no longer there, as a rename not yet added
// leaves behind, holds nothing and is left out too.
func trackedTextFiles(t *testing.T) []string {
	t.Helper()
	root := structure.Root(t)
	command := exec.Command("git", "ls-files", "-z", "--cached", "--others", "--exclude-standard")
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
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			t.Fatalf("reading %s: %v", name, err)
		}
		if !bytes.Contains(raw, []byte{0}) {
			out = append(out, name)
		}
	}
	return out
}

// stringMethod is the retired word as Python's string method, with its left and right forms, which
// is ordinary code rather than a name: a rename once turned one into a call that does not exist and
// broke the version stamp, so these calls are left alone rather than forbidden.
var stringMethod = regexp.MustCompile(`\.(l|r)?` + retiredWord + `\(`)

// holdsRetiredWord reports whether text holds the retired word in any case, outside a call of the
// string method.
func holdsRetiredWord(text string) bool {
	return strings.Contains(stringMethod.ReplaceAllString(strings.ToLower(text), ""), retiredWord)
}

func TestNoTrackedFileHoldsTheRetiredWord(t *testing.T) {
	root := structure.Root(t)
	files := trackedTextFiles(t)
	if len(files) == 0 {
		t.Fatal("git listed no tracked text files, so nothing was checked")
	}
	for _, name := range files {
		if slices.Contains(thirdPartyFiles, name) {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatalf("reading %s: %v", name, err)
		}
		for number, line := range strings.Split(string(raw), "\n") {
			if holdsRetiredWord(line) {
				t.Errorf("%s:%d holds the retired word", name, number+1)
			}
		}
	}
}

func TestTheRetiredWordIsFoundInAnyCaseAndInsideNames(t *testing.T) {
	for _, text := range []string{
		"Time" + "St" + "rip", "time" + "st" + "rip.exe", "TIME" + "ST" + "RIP_RUNLOG",
		"Show " + "st" + "rip", "find" + "St" + "rip", "." + "st" + "rip {",
	} {
		if !holdsRetiredWord(text) {
			t.Errorf("%q was not recognised", text)
		}
	}
	if holdsRetiredWord("TimeRibbon shows a ribbon of clocks") {
		t.Error("the current names were taken for the retired word")
	}
	for _, code := range []string{"text." + "st" + "rip()", "line.l" + "st" + "rip()", "name.r" + "st" + "rip('/')"} {
		if holdsRetiredWord(code) {
			t.Errorf("the string method in %q was taken for the retired word", code)
		}
	}
	if !holdsRetiredWord("text." + "st" + "rip() then the " + "st" + "rip") {
		t.Error("the word beside a string method call was missed")
	}
}
