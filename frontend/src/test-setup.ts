import { cleanup } from '@testing-library/react'
import { afterEach } from 'vitest'

// jsdom has no PointerEvent (measured: typeof PointerEvent was 'undefined'), so a fired pointer
// event carried neither its position nor its buttons. MouseEvent carries both.
if (typeof window.PointerEvent === 'undefined') {
  class PointerEventStandIn extends MouseEvent {}
  window.PointerEvent = PointerEventStandIn as typeof PointerEvent
}

afterEach(() => {
  cleanup()
})
