import type { CSSProperties } from 'react'
import { api, type Box, type Refused, type Snapshot } from './api'
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
}

/**
 * Surface is the window's content while it is the ribbon: the ribbon alone, else the ribbon with its
 * sun map beside it; a vertical ribbon also carries the handle that pulls the map out (FR-902, FR-903).
 * Go places the two in window pixels; the window's own width turns them into the page's units.
 */
export function Surface({ snapshot, onAddClock, refused }: Props) {
  const map = snapshot.sunMap
  if (snapshot.collapsed || map.side === '') {
    return <Ribbon snapshot={snapshot} onAddClock={onAddClock} refused={refused} />
  }
  const right = Math.max(map.ribbon.x + map.ribbon.width, map.shown ? map.map.x + map.map.width : 0)
  const scale = right > 0 ? window.innerWidth / right : 1
  const place = (box: Box): CSSProperties => ({ left: box.x * scale, top: box.y * scale, width: box.width * scale, height: box.height * scale })
  const ribbon = place(map.ribbon)
  const arrow = arrows[map.side]
  return (
    <div className="surface">
      <div className="surface-part" style={ribbon}>
        <Ribbon snapshot={snapshot} onAddClock={onAddClock} refused={refused} />
      </div>
      {map.shown && (
        <div className="surface-part" style={place(map.map)}>
          <SunMap sunMap={map} width={map.map.width * scale} height={map.map.height * scale} dragThreshold={snapshot.dragThreshold} refused={refused} />
        </div>
      )}
      {arrow !== undefined && (
        <button
          type="button"
          className={`pull-out-handle ${map.side}`}
          style={{ top: Number(ribbon.top) + Number(ribbon.height) / 2, left: map.side === 'left' ? ribbon.left : Number(ribbon.left) + Number(ribbon.width) }}
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
