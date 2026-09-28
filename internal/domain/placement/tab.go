package placement

// TabThickness is how deep, in DIP, the tab of an unpinned ribbon is (FR-614).
const TabThickness = 8

// Tab answers the rectangle an unpinned ribbon at at of size shrinks to on work (FR-614): a band
// thickness pixels deep along the ribbon's whole length, covering the side nearer the matching edge
// of work. That side is left or right for a vertical ribbon, top or bottom for a horizontal one; at
// an equal distance it is the side of home, the orientation's home edge (FR-409).
func Tab(at Point, size Size, work Rect, vertical bool, home Edge, thickness int) Rect {
	ribbon := Rect{Left: at.X, Top: at.Y, Right: at.X + size.Width, Bottom: at.Y + size.Height}
	if vertical {
		before, after := ribbon.Left-work.Left, work.Right-ribbon.Right
		if before < after || before == after && home == Left {
			ribbon.Right = ribbon.Left + thickness
		} else {
			ribbon.Left = ribbon.Right - thickness
		}
		return ribbon
	}
	before, after := ribbon.Top-work.Top, work.Bottom-ribbon.Bottom
	if before < after || before == after && home != Bottom {
		ribbon.Bottom = ribbon.Top + thickness
	} else {
		ribbon.Top = ribbon.Bottom - thickness
	}
	return ribbon
}
