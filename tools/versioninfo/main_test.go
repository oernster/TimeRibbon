package main

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/oernster/timeribbon/internal/product"
)

func TestTheResourceCarriesTheVersionAndTheIdentity(t *testing.T) {
	t.Parallel()
	for _, setup := range []bool{false, true} {
		body, err := describe("1.2.3", setup)
		if err != nil {
			t.Fatal(err)
		}
		var r resource
		if err := json.Unmarshal(body, &r); err != nil {
			t.Fatal(err)
		}
		strings := r.Info[neutralLanguage]
		wantDescription := product.Name
		if setup {
			wantDescription += setupSuffix
		}
		if r.Fixed.FileVersion != "1.2.3" || r.Fixed.ProductVersion != "1.2.3" || strings["ProductVersion"] != "1.2.3" ||
			strings["FileVersion"] != "1.2.3" || strings["ProductName"] != product.Name ||
			strings["FileDescription"] != wantDescription || strings["CompanyName"] != product.Author ||
			strings["LegalCopyright"] != product.Copyright {
			t.Errorf("setup %v: %s", setup, body)
		}
	}
}

func TestAVersionThatIsNotMajorMinorPatchIsRefused(t *testing.T) {
	t.Parallel()
	for _, bad := range []string{"", "1.0", "1.0.0-dev", "v1.0.0"} {
		if _, err := describe(bad, false); !errors.Is(err, ErrBadVersion) {
			t.Errorf("%q: got %v", bad, err)
		}
	}
}

func TestRunWritesTheFileMakingItsFolder(t *testing.T) {
	t.Parallel()
	out := filepath.Join(t.TempDir(), "windows", "info.json")
	if err := run([]string{"-version", "1.0.0", "-setup", "-out", out}, io.Discard); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var r resource
	if err := json.Unmarshal(raw, &r); err != nil || r.Info[neutralLanguage]["FileDescription"] != product.Name+setupSuffix {
		t.Errorf("%v: %s", err, raw)
	}
}

func TestRunRefusesWhatItCannotUse(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cases := map[string][]string{
		"no -out":           {"-version", "1.0.0"},
		"a bad version":     {"-version", "one", "-out", filepath.Join(dir, "a.json")},
		"an unknown flag":   {"-colour", "blue"},
		"an unwritable out": {"-version", "1.0.0", "-out", dir},
	}
	for name, args := range cases {
		if err := run(args, io.Discard); err == nil {
			t.Errorf("%s was accepted", name)
		}
	}
}
