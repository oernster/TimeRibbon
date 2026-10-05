package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"

	"github.com/oernster/ribbonkit/infrastructure/settingsfile"
	"github.com/oernster/timeribbon/internal/domain/settings"
)

// formatVersion is written into every file so a later version can tell what it is reading.
const formatVersion = 1

// TimeRibbon's own top-level keys; the ribbon's are the kit's.
const (
	keyVersion    = "version"
	keyStyle      = "style"
	keySize       = "size"
	keyFormat     = "format"
	keyClocks     = "clocks"
	keyDateFormat = "dateFormat"
	keySunMap     = "sunMap"
	keyPullOut    = "pullOut"
)

// knownKeys lists the keys this version reads, in writing order. Every key from skippedUpdate on came
// after 1.0.0, so they are written last (NFR-C-1).
var knownKeys = []string{
	keyVersion, keyStyle, keySize, settingsfile.KeyColour, keyFormat, settingsfile.KeyOrientation,
	settingsfile.KeyTheme, settingsfile.KeyAlwaysOnTop, settingsfile.KeyPlacement, keyClocks,
	settingsfile.KeySkippedUpdate, keyDateFormat, settingsfile.KeyPinned, settingsfile.KeyLastEdge, keySunMap,
	keyPullOut, settingsfile.KeyOpacity, settingsfile.KeyScale, settingsfile.KeyPullOutSide,
}

// codec is TimeRibbon's settings file.
var codec = settingsfile.Codec[settings.Settings]{
	Keys:     knownKeys,
	Defaults: settings.Defaults,
	Decode:   decode,
	Values:   values,
}

// Why a clock entry cannot be read, as the user is shown it (FR-705).
var (
	errNoID   = errors.New("it has no id")
	errNoZone = errors.New("it names no time zone")
)

// storedClock is one clock as the file holds it. Pointers tell a missing field from an empty one.
type storedClock struct {
	ID       *string `json:"id"`
	Zone     *string `json:"zone"`
	Label    *string `json:"label"`
	Position *int    `json:"position"`
}

// clocks reads each clock entry on its own, so one bad entry becomes an unreadable clock and the rest
// load (FR-705).
var clocks = settingsfile.Entries[settings.Entry]{
	Decode: decodeClock,
	Encode: func(entry settings.Entry, position int) (json.RawMessage, error) {
		return json.Marshal(storedClock{ID: &entry.ID, Zone: &entry.Zone, Label: &entry.Label, Position: &position})
	},
	Unreadable: func(id, reason, original string) settings.Entry {
		return settings.Entry{ID: id, Unreadable: reason, Original: original}
	},
	ID:       func(entry settings.Entry) (string, bool) { return entry.ID, entry.Unreadable == "" },
	WithID:   func(entry settings.Entry, id string) settings.Entry { entry.ID = id; return entry },
	Original: func(entry settings.Entry) string { return entry.Original },
}

// decode reads a settings file's object. It answers false when the clocks are not a list, since then
// nothing can be trusted. Any other bad value leaves its default.
func decode(object settingsfile.Object) (settings.Settings, bool) {
	entries, ok := object.List(keyClocks)
	if !ok {
		return settings.Settings{}, false
	}
	decoded := settings.Defaults()
	settingsfile.ReadChoices(object, &decoded.Choices)
	settingsfile.Read(object, keyStyle, &decoded.Style)
	settingsfile.Read(object, keySize, &decoded.Size)
	settingsfile.Read(object, keyFormat, &decoded.Format)
	settingsfile.Read(object, keyDateFormat, &decoded.DateFormat)
	settingsfile.Read(object, keySunMap, &decoded.SunMap)
	settingsfile.Read(object, keyPullOut, &decoded.PullOut)
	decoded.Clocks = clocks.Read(entries)
	return decoded, true
}

// decodeClock reads one entry with its stored position; the reason where it cannot be read.
func decodeClock(raw json.RawMessage) (settings.Entry, *int, error) {
	var stored storedClock
	if err := json.Unmarshal(raw, &stored); err != nil {
		return settings.Entry{}, nil, fmt.Errorf("its fields are not what a clock holds (%v)", err)
	}
	switch {
	case stored.ID == nil || *stored.ID == "":
		return settings.Entry{}, nil, errNoID
	case stored.Zone == nil:
		return settings.Entry{}, nil, errNoZone
	}
	entry := settings.Entry{ID: *stored.ID, Zone: *stored.Zone}
	if stored.Label != nil {
		entry.Label = *stored.Label
	}
	return entry, stored.Position, nil
}

// values answers the value of every key, with any failure writing the clocks, which refuses the save.
// Derived values are never written (FR-701).
func values(current settings.Settings) (map[string]any, error) {
	written, err := clocks.Write(current.Clocks)
	all := map[string]any{
		keyVersion: formatVersion, keyStyle: current.Style, keySize: current.Size, keyFormat: current.Format,
		keyClocks: written, keyDateFormat: current.DateFormat, keySunMap: current.SunMap, keyPullOut: current.PullOut,
	}
	maps.Copy(all, settingsfile.ChoiceValues(current.Choices))
	return all, err
}
