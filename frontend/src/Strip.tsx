import { useRef, type CSSProperties, type PointerEvent, type WheelEvent } from 'react'
import { api, startDrag, type Refused, type Snapshot } from './api'
import { ArtButton, addClockTip } from './ArtButton'
import addClockArt from './assets/add-clock.png'
import { Cell } from './Cell'

interface Props {
  snapshot: Snapshot
  onAddClock: () => void
  refused: Refused
}

/** Presses on these start no drag (FR-402). */
const controls = 'button, input, select, a, [data-control]'

/**
 * Strip is the clocks in order (FR-102). Pressing empty strip area and moving past Windows' drag
 * distance moves the whole window (FR-401); a small wobble or a press on a control does not.
 */
export function Strip({ snapshot, onAddClock, refused }: Props) {
  const pressed = useRef<{ x: number; y: number } | null>(null)
  const vertical = snapshot.orientation === 'vertical'
  const analogue = snapshot.style === 'analogue'
  const cell = snapshot.cells.length === 0 ? snapshot.layout.prompt : analogue ? snapshot.layout.analogue : snapshot.layout.digital
  const sizing = {
    '--cell-w': `${cell.width}px`,
    '--cell-h': `${cell.height}px`,
    '--pad': `${snapshot.layout.padding}px`,
  } as CSSProperties

  const down = (event: PointerEvent<HTMLDivElement>) => {
    const target = event.target as HTMLElement
    pressed.current = event.button === 0 && target.closest(controls) == null ? { x: event.screenX, y: event.screenY } : null
  }
  const move = (event: PointerEvent<HTMLDivElement>) => {
    const start = pressed.current
    if (start == null || event.buttons !== 1) {
      return
    }
    const threshold = snapshot.dragThreshold
    if (Math.abs(event.screenX - start.x) > threshold.width || Math.abs(event.screenY - start.y) > threshold.height) {
      pressed.current = null
      startDrag()
    }
  }
  const up = () => {
    pressed.current = null
  }
  // A plain wheel moves up and down, which a horizontal strip cannot; so while one scrolls, the
  // wheel moves it along instead (FR-106). A sideways wheel or a trackpad already moves it along.
  const wheel = (event: WheelEvent<HTMLDivElement>) => {
    if (!vertical && snapshot.scrolls && event.deltaX === 0) {
      event.currentTarget.scrollLeft += event.deltaY
    }
  }

  const classes = ['strip', vertical ? 'vertical' : 'horizontal', snapshot.size, snapshot.scrolls ? 'scrolls' : ''].join(' ')
  return (
    <div
      className={classes}
      style={sizing}
      onPointerDown={down}
      onPointerMove={move}
      onPointerUp={up}
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
