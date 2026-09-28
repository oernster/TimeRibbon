// Command identity prints the names the Linux build needs, as shell assignments read from the one
// home each has in the code, so no build script keeps a second copy:
//
//	eval "$(go run ./tools/identity)"
package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/oernster/timeribbon/internal/product"
)

// Wails' single-instance lock on Linux owns a bus name made from the instance id, which launch.go
// sets to the app id: "org.wails_app_" then the id with dashes and dots as underscores, then
// ".SingleInstance" (read in Wails v2.12.0, internal/frontend/desktop/linux/single_instance.go).
const (
	singleInstancePrefix = "org.wails_app_"
	singleInstanceSuffix = ".SingleInstance"
)

func main() { print(os.Stdout) }

// print writes each name as NAME='value'.
func print(out io.Writer) {
	for _, pair := range [][2]string{
		{"APP_NAME", product.Name},
		{"APP_ID", product.AppID},
		{"BIN_NAME", strings.ToLower(product.Name)},
		{"SINGLE_INSTANCE_NAME", singleInstanceName(product.AppID)},
	} {
		fmt.Fprintf(out, "%s='%s'\n", pair[0], pair[1])
	}
}

// singleInstanceName answers the bus name Wails' lock owns for id.
func singleInstanceName(id string) string {
	return singleInstancePrefix + strings.NewReplacer("-", "_", ".", "_").Replace(id) + singleInstanceSuffix
}
