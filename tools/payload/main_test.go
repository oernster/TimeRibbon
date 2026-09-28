//go:build windows

package main

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/oernster/timeribbon/internal/infrastructure/setup"
)

// put writes body at path.
func put(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), filePerm); err != nil {
		t.Fatal(err)
	}
}

func TestThePayloadHoldsTheApplicationAndItsLicence(t *testing.T) {
	t.Parallel()
	app, dir := t.TempDir(), t.TempDir()
	put(t, filepath.Join(app, setup.ExeName), "the program")
	licence := filepath.Join(dir, "LICENSE")
	put(t, licence, "terms")
	archive := filepath.Join(dir, "payload.zip")
	if err := run([]string{"-app", app, "-licence", licence, "-out", archive}, io.Discard); err != nil {
		t.Fatal(err)
	}
	reader, err := zip.OpenReader(archive)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	if len(reader.File) != 2 || reader.File[0].Name != setup.ExeName || reader.File[1].Name != setup.LicenceFile {
		t.Errorf("the archive holds %v", reader.File)
	}
	if _, err := os.Stat(archive + packingSuffix); err == nil {
		t.Error("the packing file was left behind")
	}
}

// A packing that is refused leaves the archive that was there before.
func TestARefusedPackingLeavesTheOldArchive(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	archive := filepath.Join(dir, "payload.zip")
	put(t, archive, "the placeholder")
	err := run([]string{"-app", t.TempDir(), "-licence", "x", "-out", archive}, io.Discard)
	if !errors.Is(err, setup.ErrNoApplication) {
		t.Errorf("got %v", err)
	}
	if got, _ := os.ReadFile(archive); string(got) != "the placeholder" {
		t.Errorf("the archive now holds %q", got)
	}
	if _, err := os.Stat(archive + packingSuffix); err == nil {
		t.Error("the packing file was left behind")
	}
}

func TestEveryFlagIsNeeded(t *testing.T) {
	t.Parallel()
	if err := run([]string{"-app", "a", "-out", "b"}, io.Discard); !errors.Is(err, errMissingFlag) {
		t.Errorf("got %v", err)
	}
	if err := run([]string{"-unknown"}, io.Discard); err == nil {
		t.Error("an unknown flag was taken")
	}
	if err := run([]string{"-app", "a", "-licence", "l", "-out", filepath.Join(t.TempDir(), "absent", "p.zip")}, io.Discard); err == nil {
		t.Error("an archive in a folder that is not there was written")
	}
}
