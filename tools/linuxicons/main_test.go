package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// Every theme size is written from the master, each at its own size.
func TestEverySizeIsWritten(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	master := filepath.Join(dir, "master.png")
	file, err := os.Create(master)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(file, image.NewNRGBA(image.Rect(0, 0, 8, 8))); err != nil {
		t.Fatal(err)
	}
	_ = file.Close()
	out := filepath.Join(dir, "icons")
	if err := run([]string{"-in", master, "-out", out, "-name", "ribbon"}); err != nil {
		t.Fatal(err)
	}
	for _, size := range themeSizes {
		written, err := os.Open(filepath.Join(out, fmt.Sprintf("ribbon_%d.png", size)))
		if err != nil {
			t.Fatal(err)
		}
		config, err := png.DecodeConfig(written)
		_ = written.Close()
		if err != nil || config.Width != size || config.Height != size {
			t.Errorf("size %d: %dx%d (%v)", size, config.Width, config.Height, err)
		}
	}
}

func TestMissingArgumentsAndAMissingMasterAreRefused(t *testing.T) {
	t.Parallel()
	if err := run([]string{"-in", "x.png"}); err == nil {
		t.Error("run without -out and -name was not refused")
	}
	missing := filepath.Join(t.TempDir(), "absent.png")
	if err := run([]string{"-in", missing, "-out", t.TempDir(), "-name", "r"}); err == nil {
		t.Error("a missing master was not refused")
	}
}
