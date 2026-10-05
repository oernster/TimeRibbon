package occupancy

import (
	"github.com/oernster/timeribbon/ribbonkit/domain/identity"
	"github.com/oernster/timeribbon/ribbonkit/infrastructure/appdata"
)

// dir answers ~/Library/Application Support/ribbonkit: the folder appdata names for an application
// called after the kit.
func dir(lookup func(string) (string, bool)) (string, error) {
	folder, err := appdata.Dir(identity.App{Name: kitFolder}, lookup)
	if err != nil {
		return "", ErrNoFolder
	}
	return folder, nil
}
