package placement

// The sun map's shape and place beside the ribbon (FR-902 to FR-904). The map is a whole-world
// picture twice as wide as it is deep.

const (
	// MapMinimumWidth is the narrowest, in DIP, the sun map is drawn; its depth is half (FR-904).
	MapMinimumWidth = 480
	// MapFloor is the least room, in DIP, beside the ribbon that the sun map is shown in (FR-904).
	MapFloor = 120
	// mapAspect is how many times wider than deep the world map is.
	mapAspect = 2
)

// InnerSide answers the ribbon's long side facing away from the edge it stands against, where the
// sun map and the pull out's handle go (FR-902, FR-903): below or above a horizontal ribbon, left or
// right of a vertical one. Against no edge it is the side facing the more room in work; at an equal
// room, below a horizontal ribbon and left of a vertical one, the side away from each home edge.
func InnerSide(ribbon Rect, work Rect, vertical bool, edge Edge) Edge {
	switch edge {
	case Top:
		return Bottom
	case Bottom:
		return Top
	case Left:
		return Right
	case Right:
		return Left
	}
	if vertical {
		if work.Right-ribbon.Right > ribbon.Left-work.Left {
			return Right
		}
		return Left
	}
	if ribbon.Top-work.Top > work.Bottom-ribbon.Bottom {
		return Top
	}
	return Bottom
}

// MapBeside answers the sun map's rectangle adjoining side of a ribbon on work (FR-904): as wide as
// the ribbon is long, never narrower than minimumWidth, half as deep as it is wide; scaled down,
// keeping its shape, to the room beside the ribbon; centred on the ribbon along it and kept inside
// work. It answers false where that room is less than floor, when no map is shown. minimumWidth and
// floor are in the display's pixels.
func MapBeside(ribbon Rect, work Rect, side Edge, minimumWidth, floor int) (Rect, bool) {
	vertical := side == Left || side == Right
	room := roomBeside(ribbon, work, side)
	if room < floor {
		return Rect{}, false
	}
	length := ribbon.Width()
	if vertical {
		length = ribbon.Height()
	}
	width := min(max(length, minimumWidth), work.Width())
	if vertical {
		width = min(width, room)
	} else {
		width = min(width, room*mapAspect)
	}
	height := width / mapAspect
	size := Size{Width: width, Height: height}
	var at Point
	switch side {
	case Bottom:
		at = Point{X: centreOn(ribbon.Left, ribbon.Width(), width), Y: ribbon.Bottom}
	case Top:
		at = Point{X: centreOn(ribbon.Left, ribbon.Width(), width), Y: ribbon.Top - height}
	case Right:
		at = Point{X: ribbon.Right, Y: centreOn(ribbon.Top, ribbon.Height(), height)}
	default:
		at = Point{X: ribbon.Left - width, Y: centreOn(ribbon.Top, ribbon.Height(), height)}
	}
	return rectOf(Clamp(at, size, work), size), true
}

// roomBeside answers the room in work beyond the ribbon's side.
func roomBeside(ribbon Rect, work Rect, side Edge) int {
	switch side {
	case Bottom:
		return work.Bottom - ribbon.Bottom
	case Top:
		return ribbon.Top - work.Top
	case Right:
		return work.Right - ribbon.Right
	}
	return ribbon.Left - work.Left
}

// centreOn answers where a length starts to be centred on a span starting at start.
func centreOn(start, span, length int) int { return start + (span-length)/2 }
