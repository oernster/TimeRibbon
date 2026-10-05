import { fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { installBridge, snapshot } from './fakeBridge'
import { Ribbon, scaleGripTip } from './Ribbon'

// The ribbon the grip sits in, in the page's pixels across it. The grip's own behaviour is
// ribbonkit's (ScaleGrip.test.tsx); these are the ribbon's use of it.
const thickness = 100

afterEach(() => vi.restoreAllMocks())

describe('the ribbon and its grip (FR-623)', () => {
  function grip() {
    const element = screen.getByRole('separator', { name: scaleGripTip })
    vi.spyOn(element.parentElement as HTMLElement, 'getBoundingClientRect').mockReturnValue({ width: thickness * 2, height: thickness } as DOMRect)
    return element
  }

  it('draws the clocks at the snapshot scale', () => {
    installBridge()
    render(<Ribbon snapshot={snapshot({ scale: 120 })} onAddClock={vi.fn()} refused={vi.fn()} />)
    const ribbon = document.querySelector('.ribbon') as HTMLElement
    expect(ribbon.style.zoom).toBe('1.2')
  })

  it('measures a vertical ribbon across its width', () => {
    const bridge = installBridge()
    render(<Ribbon snapshot={snapshot({ orientation: 'vertical' })} onAddClock={vi.fn()} refused={vi.fn()} />)
    fireEvent.pointerDown(grip(), { button: 0, screenX: 0, screenY: 0 })
    expect(bridge.BeginScale).toHaveBeenCalledWith(thickness * 2, 0, 0)
  })

  it('is a control, so pressing it starts no window drag (FR-402)', () => {
    installBridge()
    render(<Ribbon snapshot={snapshot()} onAddClock={vi.fn()} refused={vi.fn()} />)
    const handle = grip()
    fireEvent.pointerDown(handle, { button: 0, screenX: 0, screenY: 0 })
    fireEvent.pointerMove(handle, { buttons: 1, screenX: 50, screenY: 50 })
    expect(window.WailsInvoke).not.toHaveBeenCalled()
  })
})
