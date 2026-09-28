package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"

	"github.com/oernster/timeribbon/internal/domain/placement"
	"github.com/oernster/timeribbon/internal/domain/settings"
)

// Top-level keys, in the order they are written.
const (
	keyVersion     = "version"
	keyStyle       = "style"
	keySize        = "size"
	keyColour      = "colour"
	keyFormat      = "format"
	keyOrientation = "orientation"
	keyTheme       = "theme"
	keyAlwaysOnTop = "alwaysOnTop"
	keyPlacement   = "placement"
	keyClocks      = "clocks"
	// keySkippedUpdate came after 1.0.0, so it is written last (NFR-C-1).
	keySkippedUpdate = "skippedUpdate"
)

// knownKeys lists the keys this version reads, in writing order.
var knownKeys = []string{
	keyVersion, keyStyle, keySize, keyColour, keyFormat, keyOrientation, keyTheme, keyAlwaysOnTop, keyPlacement, keyClocks,
	keySkippedUpdate,
}

// unreadableIDPrefix begins the id an unreadable clock is given for the session, so it can be
// edited or removed. It is never written: the entry is written back as it was found.
const unreadableIDPrefix = "unreadable-"

// storedClock is one clock as the file holds it. Pointers tell a missing field from an empty one.
type storedClock struct {
	ID       *string `json:"id"`
	Zone     *string `json:"zone"`
	Label    *string `json:"label"`
	Position *int    `json:"position"`
}

// storedPlacement is the placement as the file holds it.
type storedPlacement struct {
	Device string `json:"device"`
	Work   struct {
		Left   int `json:"left"`
		Top    int `json:"top"`
		Right  int `json:"right"`
		Bottom int `json:"bottom"`
	} `json:"work"`
	DPI    int `json:"dpi"`
	Offset struct {
		X int `json:"x"`
		Y int `json:"y"`
	} `json:"offset"`
}

// decode reads a settings file. It answers false when the file is not a JSON object or its clocks
// are not a list, since then nothing can be trusted. Any other bad value leaves its default.
func decode(raw []byte) (settings.Settings, []pair, bool) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil || object == nil {
		return settings.Settings{}, nil, false
	}
	var entries []json.RawMessage
	if list, ok := object[keyClocks]; ok && string(list) != "null" {
		if err := json.Unmarshal(list, &entries); err != nil {
			return settings.Settings{}, nil, false
		}
	}
	decoded := settings.Defaults()
	readInto(object, keyStyle, &decoded.Style)
	readInto(object, keySize, &decoded.Size)
	readInto(object, keyColour, &decoded.Colour)
	readInto(object, keyFormat, &decoded.Format)
	readInto(object, keyOrientation, &decoded.Orientation)
	readInto(object, keyTheme, &decoded.Theme)
	readInto(object, keyAlwaysOnTop, &decoded.AlwaysOnTop)
	readInto(object, keySkippedUpdate, &decoded.SkippedUpdate)
	decoded.Placement = decodePlacement(object[keyPlacement])
	decoded.Clocks = decodeClocks(entries)
	return decoded, extrasOf(raw, object), true
}

// readInto decodes object[key] into target, leaving target as it was when the key is missing or
// its value is the wrong type.
func readInto[T any](object map[string]json.RawMessage, key string, target *T) {
	value, ok := object[key]
	if !ok {
		return
	}
	var read T
	if json.Unmarshal(value, &read) == nil {
		*target = read
	}
}

func decodePlacement(raw json.RawMessage) *placement.Stored {
	if len(raw) == 0 {
		return nil
	}
	var stored storedPlacement
	if json.Unmarshal(raw, &stored) != nil || stored.Device == "" {
		return nil
	}
	return &placement.Stored{
		Device: stored.Device,
		Work:   placement.Rect{Left: stored.Work.Left, Top: stored.Work.Top, Right: stored.Work.Right, Bottom: stored.Work.Bottom},
		DPI:    stored.DPI,
		Offset: placement.Point{X: stored.Offset.X, Y: stored.Offset.Y},
	}
}

// decodeClocks reads each entry on its own, so one bad entry becomes an unreadable clock and the
// rest load (FR-705). Entries are ordered by their stored position; an entry with none keeps its
// place in the list.
func decodeClocks(entries []json.RawMessage) []settings.Entry {
	type ordered struct {
		entry    settings.Entry
		position int
	}
	read := make([]ordered, 0, len(entries))
	for index, raw := range entries {
		entry, position := decodeClock(raw, index)
		read = append(read, ordered{entry: entry, position: position})
	}
	slices.SortStableFunc(read, func(a, b ordered) int { return a.position - b.position })
	out := make([]settings.Entry, 0, len(read))
	for _, each := range read {
		out = append(out, each.entry)
	}
	return out
}

// decodeClock reads one entry, answering it with its position; its index when it has none.
func decodeClock(raw json.RawMessage, index int) (settings.Entry, int) {
	unreadable := func(reason string) (settings.Entry, int) {
		return settings.Entry{
			ID:         unreadableIDPrefix + strconv.Itoa(index),
			Unreadable: reason,
			Original:   string(raw),
		}, index
	}
	var stored storedClock
	if err := json.Unmarshal(raw, &stored); err != nil {
		return unreadable(fmt.Sprintf("its fields are not what a clock holds (%v)", err))
	}
	switch {
	case stored.ID == nil || *stored.ID == "":
		return unreadable("it has no id")
	case stored.Zone == nil:
		return unreadable("it names no time zone")
	}
	entry := settings.Entry{ID: *stored.ID, Zone: *stored.Zone}
	if stored.Label != nil {
		entry.Label = *stored.Label
	}
	position := index
	if stored.Position != nil {
		position = *stored.Position
	}
	return entry, position
}

// extrasOf answers the top-level keys this version does not know, in the order the file has them.
func extrasOf(raw []byte, object map[string]json.RawMessage) []pair {
	var extras []pair
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if _, err := decoder.Token(); err != nil {
		return nil
	}
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return extras
		}
		key, _ := token.(string)
		var skipped json.RawMessage
		if decoder.Decode(&skipped) != nil {
			return extras
		}
		if !slices.Contains(knownKeys, key) {
			extras = append(extras, pair{key: key, value: object[key]})
		}
	}
	return extras
}
