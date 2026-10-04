import type { CSSProperties, MouseEvent, PointerEvent } from 'react'
import { api, type Box, type Refused, type Snapshot } from './api'
import { showsTheMenu, useDrag } from './drag'
import { percentOfWhole } from './opacity'
import { Ribbon } from './Ribbon'
import { SunMap } from './SunMap'

interface Props {
  snapshot: Snapshot
  onAddClock: () => void
  refused: Refused
}

/** The handle's words, one home: the button says what pressing it will do (NFR-U-4). */
export const openPullOut = 'Open the sun map'
export const closePullOut = 'Close the sun map'

/** The handle's arrows point the way the map will move: out towards side when closed, back when open. */
const arrows: Record<string, { out: string; back: string }> = {
  left: { out: '◀', back: '▶' },
  right: { out: '▶', back: '◀' },
  top: { out: '▲', back: '▼' },
  bottom: { out: '▼', back: '▲' },
}

/** A box in the page's units. */
interface Placed {
  left: number
  top: number
  width: number
  height: number
}

/**
 * handleAt answers where the handle is anchored: on the ribbon's side, half way along it. The
 * handle's style for that side then draws it just inside the ribbon.
 */
function handleAt(side: string, ribbon: Placed): { left: number; top: number } {
  const middle = { left: ribbon.left + ribbon.width / 2, top: ribbon.top + ribbon.height / 2 }
  switch (side) {
    case 'left':
      return { left: ribbon.left, top: middle.top }
    case 'right':
      return { left: ribbon.left + ribbon.width, top: middle.top }
    case 'top':
      return { left: middle.left, top: ribbon.top }
    default:
      return { left: middle.left, top: ribbon.top + ribbon.height }
  }
}

/**
 * Surface is the window's content while it is the ribbon: the ribbon alone, else the ribbon with its
 * sun map beside it, horizontal or vertical, with the handle that pulls the map out and puts it back
 * (FR-902, FR-903). Go places the two, in the page's units.
 */
export function Surface({ snapshot, onAddClock, refused }: Props) {
  const drag = useDrag(snapshot.dragThreshold)
  const map = snapshot.sunMap
  if (snapshot.collapsed || map.side === '') {
    return <Ribbon snapshot={snapshot} onAddClock={onAddClock} refused={refused} />
  }
  // Go sends each part already in the page's units, so it is drawn as it comes. Scaling it by the
  // window's width drew an opening ribbon, which is drawn while the window is still its tab (FR-615),
  // 8 wide in a window then grown to 175, blank and deaf to a right-click (measured 2026-09-29).
  const place = (box: Box): Placed => ({ left: box.x, top: box.y, width: box.width, height: box.height })
  const ribbon = place(map.ribbon)
  const arrow = arrows[map.side]
  // The handle stands in the lane, which is drawn at the ribbon's scale (FR-623).
  const drawnAt = snapshot.scale / percentOfWhole
  const lane = { '--handle-width': `${snapshot.layout.handleLane * drawnAt}px`, '--scale': drawnAt } as CSSProperties
  // On macOS and Linux the window stays a rectangle (FR-913), so beside a ribbon shorter than its map
  // the surface itself shows, painted as the ribbon is; it answers a right-click and a drag as the
  // ribbon does. A press on the ribbon or the map reaches here too, which they already answer, so
  // only one on the surface itself counts.
  const menu = showsTheMenu(refused)
  const own = {
    ...drag,
    onPointerDown: (event: PointerEvent<HTMLElement>) => {
      if (event.target === event.currentTarget) {
        drag.onPointerDown(event)
      }
    },
    onContextMenu: (event: MouseEvent<HTMLElement>) => {
      if (event.target === event.currentTarget) {
        menu(event)
      }
    },
  }
  return (
    <div className="surface" style={lane} {...own}>
      <div className="surface-part" style={ribbon}>
        <Ribbon snapshot={snapshot} onAddClock={onAddClock} refused={refused} />
      </div>
      {map.shown && (
        <div className="surface-part" style={place(map.map)}>
          <SunMap sunMap={map} width={map.map.width} height={map.map.height} dragThreshold={snapshot.dragThreshold} refused={refused} />
        </div>
      )}
      {arrow !== undefined && (
        <button
          type="button"
          className={`pull-out-handle ${map.side}`}
          style={handleAt(map.side, ribbon)}
          aria-label={map.pullOut ? closePullOut : openPullOut}
          title={map.pullOut ? closePullOut : openPullOut}
          onClick={() => void api.togglePullOut(refused)}
        >
          {map.pullOut ? arrow.back : arrow.out}
        </button>
      )}
    </div>
  )
}
