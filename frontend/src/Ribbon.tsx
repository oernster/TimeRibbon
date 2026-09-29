import { useEffect, type CSSProperties, type WheelEvent } from 'react'
import { api, type Refused, type Snapshot } from './api'
import { ArtButton, addClockTip } from './ArtButton'
import addClockArt from './assets/add-clock.png'
import { Cell } from './Cell'
import { useDrag } from './drag'

interface Props {
  snapshot: Snapshot
  onAddClock: () => void
  refused: Refused
}

/**
 * Ribbon is the clocks in order (FR-102). Pressing empty ribbon area and moving past Windows' drag
 * distance moves the whole window (FR-401); a small wobble or a press on a control does not.
 */
export function Ribbon({ snapshot, onAddClock, refused }: Props) {
  const drag = useDrag(snapshot.dragThreshold)
  const vertical = snapshot.orientation === 'vertical'
  const analogue = snapshot.style === 'analogue'
  const cell = snapshot.cells.length === 0 ? snapshot.layout.prompt : analogue ? snapshot.layout.analogue : snapshot.layout.digital
  const sizing = {
    '--cell-w': `${cell.width}px`,
    '--cell-h': `${cell.height}px`,
    '--pad': `${snapshot.layout.padding}px`,
  } as CSSProperties

  // A plain wheel moves up and down, which a horizontal ribbon cannot; so while one scrolls, the
  // wheel moves it along instead (FR-106). A sideways wheel or a trackpad already moves it along.
  const wheel = (event: WheelEvent<HTMLDivElement>) => {
    if (!vertical && snapshot.scrolls && event.deltaX === 0) {
      event.currentTarget.scrollLeft += event.deltaY
    }
  }

  // An opening ribbon is drawn inside its tab first; Go grows the window once told it has been, so
  // the band is never seen stretched over the full window (FR-615). The second frame is the one
  // after the ribbon was painted.
  useEffect(() => {
    if (snapshot.collapsed) {
      return
    }
    let frame = window.requestAnimationFrame(() => {
      frame = window.requestAnimationFrame(() => void api.ribbonDrawn(refused))
    })
    return () => window.cancelAnimationFrame(frame)
  }, [snapshot.collapsed, refused])

  // An unpinned ribbon's tab is only a band in the scheme's accent: no words, no drag, no menu. It
  // opens when the pointer rests on it, which the desktop reports rather than the page (FR-614).
  if (snapshot.collapsed) {
    return <div className="tab" data-testid="tab" />
  }

  const classes = ['ribbon', vertical ? 'vertical' : 'horizontal', snapshot.size, snapshot.scrolls ? 'scrolls' : ''].join(' ')
  return (
    <div
      className={classes}
      style={sizing}
      {...drag}
      onWheel={wheel}
      onContextMenu={(event) => {
        event.preventDefault()
        void api.showContextMenu(refused)
      }}
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
    </div>
  )
}
