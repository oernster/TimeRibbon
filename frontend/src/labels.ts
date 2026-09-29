/** Where a label stands beside its dot on the sun map (FR-914). */
export type Spot = 'right' | 'left' | 'below' | 'above'

/** The spots in the order they are tried; right is where every label stood before FR-914. */
export const spots: readonly Spot[] = ['right', 'left', 'below', 'above']

/** A point on the map, in CSS pixels from its top-left corner. */
export interface Point {
  x: number
  y: number
}

/** A width by height, in CSS pixels. */
export interface Extent {
  width: number
  height: number
}

interface Box {
  left: number
  top: number
  right: number
  bottom: number
}

/** boxAt answers the box a label of size takes at spot beside a dot at centre, reach from its centre. */
function boxAt(centre: Point, size: Extent, spot: Spot, reach: number): Box {
  switch (spot) {
    case 'right':
      return { left: centre.x + reach, right: centre.x + reach + size.width, top: centre.y - size.height / 2, bottom: centre.y + size.height / 2 }
    case 'left':
      return { left: centre.x - reach - size.width, right: centre.x - reach, top: centre.y - size.height / 2, bottom: centre.y + size.height / 2 }
    case 'below':
      return { left: centre.x - size.width / 2, right: centre.x + size.width / 2, top: centre.y + reach, bottom: centre.y + reach + size.height }
    case 'above':
      return { left: centre.x - size.width / 2, right: centre.x + size.width / 2, top: centre.y - reach - size.height, bottom: centre.y - reach }
  }
}

/** overlaps reports whether two boxes share any area; boxes that only touch do not. */
function overlaps(a: Box, b: Box): boolean {
  return a.left < b.right && b.left < a.right && a.top < b.bottom && b.top < a.bottom
}

/** inside reports whether box lies wholly on a map of extent. */
function inside(box: Box, map: Extent): boolean {
  return box.left >= 0 && box.top >= 0 && box.right <= map.width && box.bottom <= map.height
}

/**
 * placeLabels answers a spot for each label, in clock order (FR-914): the first of right, left,
 * below and above that lies wholly on the map and overlaps no dot and no label already placed;
 * right where none does. dots[i] is the centre of label i's dot; dotSize is a dot's width and gap the
 * space between a dot and its label, both from the page's style.
 */
export function placeLabels(dots: readonly Point[], labels: readonly Extent[], map: Extent, dotSize: number, gap: number): Spot[] {
  const half = dotSize / 2
  const taken: Box[] = dots.map((dot) => ({ left: dot.x - half, right: dot.x + half, top: dot.y - half, bottom: dot.y + half }))
  return dots.map((dot, index) => {
    const size = labels[index]
    const spot = spots.find((candidate) => {
      const box = boxAt(dot, size, candidate, half + gap)
      return inside(box, map) && !taken.some((other) => overlaps(box, other))
    }) ?? spots[0]
    taken.push(boxAt(dot, size, spot, half + gap))
    return spot
  })
}
