// Command versioninfo writes the Windows version resource wails build puts into an executable,
// from the version build.ps1 read out of VERSION and the identity in internal/product, so neither
// the version nor the author has a second home. Left to its own template, Wails wrote its fallback
// version, a placeholder copyright and its own advertisement, measured on 2026-09-27.
//
//	go run ./tools/versioninfo -version 1.0.0 -out build/windows/info.json
//	go run ./tools/versioninfo -version 1.0.0 -setup -out installer/build/windows/info.json
//
// The output is build output: build.ps1 writes it before each build.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/oernster/timeribbon/internal/product"
)

// neutralLanguage keys the string table as language neutral with Unicode text, as Wails' own
// template does; Windows reads it back under the translation 000004b0.
const neutralLanguage = "0000"

// setupSuffix names the setup program after the product it installs.
const setupSuffix = " Setup"

// folderMode is how a missing output folder is made; fileMode how the resource is written.
const (
	folderMode = 0o755
	fileMode   = 0o644
)

// semanticVersion is the form VERSION holds and build.ps1 insists on.
var semanticVersion = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)

// ErrBadVersion is answered for a version that is not major.minor.patch.
var ErrBadVersion = errors.New("the version is not major.minor.patch")

// resource is the JSON winres reads: the fixed block and one string table per language.
type resource struct {
	Fixed struct {
		FileVersion    string `json:"file_version"`
		ProductVersion string `json:"product_version"`
	} `json:"fixed"`
	Info map[string]map[string]string `json:"info"`
}

func main() {
	if err := run(os.Args[1:], os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "versioninfo:", err)
		os.Exit(1)
	}
}

// run parses args and writes the resource.
func run(args []string, errOut io.Writer) error {
	flags := flag.NewFlagSet("versioninfo", flag.ContinueOnError)
	flags.SetOutput(errOut)
	version := flags.String("version", "", "the version, major.minor.patch")
	setup := flags.Bool("setup", false, "describe the setup program rather than the application")
	out := flags.String("out", "", "the info.json to write")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *out == "" {
		return errors.New("-out is required")
	}
	body, err := describe(*version, *setup)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(*out), folderMode); err != nil {
		return fmt.Errorf("making %s: %w", filepath.Dir(*out), err)
	}
	if err := os.WriteFile(*out, body, fileMode); err != nil {
		return fmt.Errorf("writing %s: %w", *out, err)
	}
	return nil
}

// describe answers the resource for the application; for the setup program when setup is set.
func describe(version string, setup bool) ([]byte, error) {
	if !semanticVersion.MatchString(version) {
		return nil, fmt.Errorf("%w: %q", ErrBadVersion, version)
	}
	description := product.Name
	if setup {
		description += setupSuffix
	}
	var r resource
	r.Fixed.FileVersion, r.Fixed.ProductVersion = version, version
	r.Info = map[string]map[string]string{neutralLanguage: {
		"ProductName":     product.Name,
		"ProductVersion":  version,
		"FileVersion":     version,
		"FileDescription": description,
		"CompanyName":     product.Author,
		"LegalCopyright":  product.Copyright,
	}}
	// A struct of strings and a map of strings always encodes, so there is no error to answer.
	body, _ := json.MarshalIndent(r, "", "\t")
	return append(body, '\n'), nil
}
