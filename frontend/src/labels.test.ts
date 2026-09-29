import { describe, expect, it } from 'vitest'
import { placeLabels } from './labels'
import { project } from './sunLight'

// A map 708 by 354, the pull out beside a 708-long vertical ribbon; the page's dot and gap.
const map = { width: 708, height: 354 }
const DOT = 8
const GAP = 4
const label = { width: 40, height: 16 }

describe('labels on the sun map (FR-914)', () => {
  it("moves London left of its dot, clear of Berlin's; Berlin stays right of its own", () => {
    const london = project(51.51, -0.13, map.width, map.height)
    const berlin = project(52.52, 13.4, map.width, map.height)
    expect(placeLabels([london, berlin], [label, label], map, DOT, GAP)).toEqual(['left', 'right'])
  })

  it('leaves a label with room to its right where it always stood', () => {
    expect(placeLabels([{ x: 100, y: 100 }], [label], map, DOT, GAP)).toEqual(['right'])
  })

  it('tries left, then below, then above as each spot is taken or runs off the map', () => {
    // At the right edge, so right runs off; a dot just left of it takes left.
    const edge = { x: 700, y: 100 }
    expect(placeLabels([edge], [label], map, DOT, GAP)).toEqual(['left'])
    const near = { x: 680, y: 100 }
    expect(placeLabels([near, { x: 660, y: 100 }], [label, label], map, DOT, GAP)[0]).toBe('below')
    const bottom = { x: 680, y: 348 }
    expect(placeLabels([bottom, { x: 660, y: 348 }], [label, label], map, DOT, GAP)[0]).toBe('above')
  })

  it('places a later label round an earlier one rather than the other way', () => {
    const first = { x: 100, y: 100 }
    const second = { x: 120, y: 100 }
    expect(placeLabels([first, second], [label, label], map, DOT, GAP)).toEqual(['left', 'right'])
  })

  it('falls back to the right where no spot is clear', () => {
    const crowd = [{ x: 4, y: 4 }, { x: 14, y: 4 }]
    const huge = { width: 800, height: 400 }
    expect(placeLabels(crowd, [huge, huge], map, DOT, GAP)).toEqual(['right', 'right'])
  })
})
