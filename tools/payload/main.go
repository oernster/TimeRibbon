//go:build windows

// Command payload packs the setup program's payload: every file of the built application, then the
// licence beside it. build.ps1 runs it from the repository root:
//
//	go run ./tools/payload -app build/bin -licence LICENSE -out installer/payload.zip
//
// The archive is written beside its target and moved into place only once whole, so a packing that
// fails leaves what was there before.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/oernster/timeribbon/internal/product"
	"github.com/oernster/timeribbon/ribbonkit/infrastructure/setup"
)

const (
	// packingSuffix follows the archive's name while it is packed.
	packingSuffix = ".packing"
	// filePerm is how the archive is written: the owner writes, everyone reads.
	filePerm = 0o644
)

// packs is the application the payload carries; setup finds its executable by the same name.
var packs = setup.Product{App: product.App()}

// errMissingFlag means the tool was not told where the application, the licence or the archive is.
var errMissingFlag = errors.New("-app, -licence and -out are all needed")

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "payload:", err)
		os.Exit(1)
	}
}

// run packs the application folder and the licence the arguments name into the archive they name.
func run(args []string, out io.Writer) error {
	flags := flag.NewFlagSet("payload", flag.ContinueOnError)
	flags.SetOutput(out)
	app := flags.String("app", "", "the folder holding the built application")
	licence := flags.String("licence", "", "the licence to carry beside it")
	archive := flags.String("out", "", "the archive to write")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *app == "" || *licence == "" || *archive == "" {
		return errMissingFlag
	}
	if err := writeWhole(*archive, setup.Payload{App: *app, Exe: packs.Exe(), Licence: *licence}); err != nil {
		return err
	}
	fmt.Fprintf(out, "%s holds the application and its licence\n", *archive)
	return nil
}

// writeWhole packs into a file beside archive, then moves it into place.
func writeWhole(archive string, payload setup.Payload) error {
	packing := archive + packingSuffix
	file, err := os.OpenFile(packing, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, filePerm)
	if err != nil {
		return fmt.Errorf("creating %s: %w", packing, err)
	}
	err = setup.Pack(file, payload)
	if closeErr := file.Close(); err == nil && closeErr != nil {
		err = fmt.Errorf("writing %s: %w", packing, closeErr)
	}
	if err != nil {
		_ = os.Remove(packing)
		return err
	}
	if err := os.Rename(packing, archive); err != nil {
		return fmt.Errorf("moving %s into place: %w", archive, err)
	}
	return nil
}
