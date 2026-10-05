import type { CSSProperties } from 'react'
import { api, type Refused, type Snapshot } from './api'
import { ArtButton, addClockTip } from './ArtButton'
import addClockArt from './assets/add-clock.png'
import { Cell } from './Cell'
import { Band, percentOfWhole } from '@oernster/ribbonkit'

/** The grip's words, one home: what dragging and double-clicking it do (NFR-U-4). */
export const scaleGripTip = 'Drag to resize the clocks; double-click for their own size'

interface Props {
  snapshot: Snapshot
  onAddClock: () => void
  refused: Refused
}

/** Ribbon is the clocks in order (FR-102), in ribbonkit's band: its tab, drag, menu and grip. */
export function Ribbon({ snapshot, onAddClock, refused }: Props) {
  const analogue = snapshot.style === 'analogue'
  const cell = snapshot.cells.length === 0 ? snapshot.layout.prompt : analogue ? snapshot.layout.analogue : snapshot.layout.digital
  const sizing = {
    '--cell-w': `${cell.width}px`,
    '--cell-h': `${cell.height}px`,
    '--pad': `${snapshot.layout.padding}px`,
    '--lane': `${snapshot.layout.handleLane}px`,
    // Everything in the clocks is drawn at the chosen scale together: text, dials, padding and
    // cells. Zoom leaves the ribbon's own 100 percent size alone, so it still fills its window,
    // which Go has sized at the same scale (FR-623; measured in Edge's engine, 2026-09-29).
    zoom: snapshot.scale / percentOfWhole,
  } as CSSProperties

  return (
    <Band
      collapsed={snapshot.collapsed}
      vertical={snapshot.orientation === 'vertical'}
      scrolls={snapshot.scrolls}
      // While the sun map is on, the pull out's handle stands in a lane of its own along the side the
      // map adjoins, which Go has already made the ribbon deep enough to hold (FR-903).
      lane={snapshot.sunMap.side}
      className={snapshot.size}
      style={sizing}
      dragThreshold={snapshot.dragThreshold}
      calls={api}
      refused={refused}
      gripTip={scaleGripTip}
    >
      {snapshot.notices.map((notice) => (
        <div key={notice} className="cell" role="alert">
          <div className="problem">{notice}</div>
          <button type="button" onClick={() => void api.dismissNotices(refused)}>
            OK
          </button>
        </div>
      ))}
      {snapshot.cells.length === 0 && (
        <div className="cell prompt">
          <div className="date">No clocks yet</div>
          <ArtButton art={addClockArt} label={addClockTip} large onClick={onAddClock} />
        </div>
      )}
      {snapshot.cells.map((each) => (
        <Cell key={each.id} cell={each} analogue={analogue} />
      ))}
    </Band>
  )
}
