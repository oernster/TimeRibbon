import { useRef, type PointerEvent } from 'react'
import { startDrag, type Size } from './api'

/** Presses on these start no drag (FR-402). */
const controls = 'button, input, select, a, [data-control]'

/**
 * useDrag answers the pointer handlers that move the whole window once a press on anything but a
 * control has moved past threshold (FR-401, FR-402). The ribbon and the sun map share them, so a
 * drag started on the map moves both (FR-909).
 */
export function useDrag(threshold: Size) {
  const pressed = useRef<{ x: number; y: number } | null>(null)
  return {
    onPointerDown: (event: PointerEvent<HTMLElement>) => {
      const target = event.target as HTMLElement
      pressed.current = event.button === 0 && target.closest(controls) == null ? { x: event.screenX, y: event.screenY } : null
    },
    onPointerMove: (event: PointerEvent<HTMLElement>) => {
      const start = pressed.current
      if (start == null || event.buttons !== 1) {
        return
      }
      if (Math.abs(event.screenX - start.x) > threshold.width || Math.abs(event.screenY - start.y) > threshold.height) {
        pressed.current = null
        startDrag()
      }
    },
    onPointerUp: () => {
      pressed.current = null
    },
  }
}
