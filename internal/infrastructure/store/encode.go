package store

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/oernster/timeribbon/internal/domain/placement"
	"github.com/oernster/timeribbon/internal/domain/settings"
)

// indent is the indentation the file is written with.
const indent = "  "

// encode answers the file's text: the known keys in writing order, then the keys a later version
// wrote, each as found. Derived values are never written (FR-701).
func encode(current settings.Settings, extras []pair) ([]byte, error) {
	clocks := make([]json.RawMessage, 0, len(current.Clocks))
	for position, entry := range current.Clocks {
		raw, err := encodeClock(entry, position)
		if err != nil {
			return nil, err
		}
		clocks = append(clocks, raw)
	}
	values := []any{
		formatVersion, current.Style, current.Size, current.Colour, current.Format, current.Orientation, current.Theme,
		current.AlwaysOnTop, encodePlacement(current.Placement), clocks, current.SkippedUpdate,
		current.DateFormat, current.Pinned,
	}
	var compact bytes.Buffer
	compact.WriteByte('{')
	for index, key := range knownKeys {
		value, err := json.Marshal(values[index])
		if err != nil {
			return nil, fmt.Errorf("writing %s: %w", key, err)
		}
		writeMember(&compact, index > 0, key, value)
	}
	for _, extra := range extras {
		writeMember(&compact, true, extra.key, extra.value)
	}
	compact.WriteByte('}')
	var out bytes.Buffer
	if err := json.Indent(&out, compact.Bytes(), "", indent); err != nil {
		return nil, fmt.Errorf("indenting the settings: %w", err)
	}
	out.WriteByte('\n')
	return out.Bytes(), nil
}

func writeMember(buffer *bytes.Buffer, comma bool, key string, value []byte) {
	if comma {
		buffer.WriteByte(',')
	}
	name, _ := json.Marshal(key)
	buffer.Write(name)
	buffer.WriteByte(':')
	buffer.Write(value)
}

// encodeClock answers one entry. An entry that could not be read is written back as it was found,
// so nothing the user wrote is lost (FR-705).
func encodeClock(entry settings.Entry, position int) (json.RawMessage, error) {
	if entry.Unreadable != "" && json.Valid([]byte(entry.Original)) {
		return json.RawMessage(entry.Original), nil
	}
	return json.Marshal(storedClock{ID: &entry.ID, Zone: &entry.Zone, Label: &entry.Label, Position: &position})
}

func encodePlacement(stored *placement.Stored) *storedPlacement {
	if stored == nil {
		return nil
	}
	var out storedPlacement
	out.Device = stored.Device
	out.Work.Left, out.Work.Top, out.Work.Right, out.Work.Bottom = stored.Work.Left, stored.Work.Top, stored.Work.Right, stored.Work.Bottom
	out.DPI = stored.DPI
	out.Offset.X, out.Offset.Y = stored.Offset.X, stored.Offset.Y
	return &out
}
