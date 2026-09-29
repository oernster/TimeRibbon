import { useRef, type PointerEvent } from 'react'
import { api, type Refused, type Snapshot } from './api'
import { percentOfWhole } from './opacity'

/** The grip's words, one home: what dragging and double-clicking it do (NFR-U-4). */
export const scaleGripTip = 'Drag to resize the clocks; double-click for their own size'

/** Where a drag of the grip began: the pointer, the scale then and the ribbon's thickness then. */
interface Start {
  x: number
  y: number
  scale: number
  thickness: number
}

interface Props {
  snapshot: Snapshot
  refused: Refused
}

/**
 * scaleAfter answers the scale a drag has reached: the scale it began at, grown or shrunk as the
 * ribbon's thickness would be by moving its far side with the pointer, held within the bounds.
 * Only movement across the ribbon counts, outward (right or down) growing it.
 */
export function scaleAfter(start: Start, x: number, y: number, vertical: boolean, bounds: { min: number; max: number }): number {
  const moved = vertical ? x - start.x : y - start.y
  const scale = Math.round((start.scale * (start.thickness + moved)) / start.thickness)
  return Math.min(Math.max(scale, bounds.min), bounds.max)
}

/**
 * ScaleGrip is the ribbon's corner grip (FR-623). Dragging it draws the clocks larger or smaller,
 * everything in them together, with the window following as it moves; the scale is kept once it is
 * let go. Double-clicking it draws them at their own size again. It is a control, so pressing it
 * starts no window drag (FR-402).
 */
export function ScaleGrip({ snapshot, refused }: Props) {
  const start = useRef<Start | null>(null)
  const reached = useRef<number | null>(null)
  const sending = useRef(false)
  const vertical = snapshot.orientation === 'vertical'
  const bounds = { min: snapshot.minScale, max: snapshot.maxScale }

  // One preview at a time: while one is on its way, only the newest scale waits to follow it.
  const preview = (scale: number) => {
    if (sending.current) {
      return
    }
    sending.current = true
    void api.previewScale(scale, refused).then(() => {
      sending.current = false
      if (reached.current != null && reached.current !== scale && start.current != null) {
        preview(reached.current)
      }
    })
  }

  const down = (event: PointerEvent<HTMLDivElement>) => {
    if (event.button !== 0) {
      return
    }
    event.currentTarget.setPointerCapture?.(event.pointerId)
    const ribbon = event.currentTarget.parentElement?.getBoundingClientRect()
    const thickness = (vertical ? ribbon?.width : ribbon?.height) ?? 0
    start.current = thickness > 0 ? { x: event.screenX, y: event.screenY, scale: snapshot.scale, thickness } : null
    reached.current = null
  }

  const move = (event: PointerEvent<HTMLDivElement>) => {
    if (start.current == null) {
      return
    }
    const scale = scaleAfter(start.current, event.screenX, event.screenY, vertical, bounds)
    if (scale !== (reached.current ?? start.current.scale)) {
      reached.current = scale
      preview(scale)
    }
  }

  const up = () => {
    const kept = reached.current
    start.current = null
    reached.current = null
    if (kept != null) {
      void api.setScale(kept, refused)
    }
  }

  return (
    <div
      className="scale-grip"
      data-control
      role="separator"
      aria-label={scaleGripTip}
      title={scaleGripTip}
      onPointerDown={down}
      onPointerMove={move}
      onPointerUp={up}
      onPointerCancel={up}
      onDoubleClick={() => void api.setScale(percentOfWhole, refused)}
    />
  )
}
