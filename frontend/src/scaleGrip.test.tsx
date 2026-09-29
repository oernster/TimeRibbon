import { fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { installBridge, snapshot } from './fakeBridge'
import { percentOfWhole } from './opacity'
import { Ribbon } from './Ribbon'
import { scaleAfter, scaleGripTip } from './ScaleGrip'

// The ribbon the grip sits in, in the page's pixels across it.
const thickness = 100

afterEach(() => vi.restoreAllMocks())

describe('scaleAfter (FR-623)', () => {
  const bounds = { min: 75, max: 200 }
  const start = { x: 0, y: 0, scale: percentOfWhole, thickness }

  it('grows as the far side moves outward with the pointer, across the ribbon only', () => {
    expect(scaleAfter(start, thickness / 2, 0, true, bounds)).toBe(150)
    expect(scaleAfter(start, 0, thickness / 2, false, bounds)).toBe(150)
    expect(scaleAfter(start, 0, thickness / 2, true, bounds)).toBe(percentOfWhole)
    expect(scaleAfter(start, -thickness / 10, 0, true, bounds)).toBe(90)
  })

  it('holds within the bounds however far the pointer goes', () => {
    expect(scaleAfter(start, thickness * 5, 0, true, bounds)).toBe(bounds.max)
    expect(scaleAfter(start, -thickness, 0, true, bounds)).toBe(bounds.min)
  })
})

describe('ScaleGrip (FR-623)', () => {
  function grip() {
    const element = screen.getByRole('separator', { name: scaleGripTip })
    vi.spyOn(element.parentElement as HTMLElement, 'getBoundingClientRect').mockReturnValue({ width: thickness, height: thickness } as DOMRect)
    return element
  }

  it('draws the clocks at the snapshot scale and previews a drag, keeping the scale once let go', async () => {
    const bridge = installBridge()
    render(<Ribbon snapshot={snapshot({ scale: 120 })} onAddClock={vi.fn()} refused={vi.fn()} />)
    const ribbon = document.querySelector('.ribbon') as HTMLElement
    expect(ribbon.style.zoom).toBe('1.2')
    const handle = grip()
    fireEvent.pointerDown(handle, { button: 0, screenX: 0, screenY: 0 })
    fireEvent.pointerMove(handle, { screenX: 0, screenY: thickness / 2 })
    await vi.waitFor(() => expect(bridge.PreviewScale).toHaveBeenCalledWith(180))
    expect(bridge.SetScale).not.toHaveBeenCalled()
    fireEvent.pointerUp(handle)
    expect(bridge.SetScale).toHaveBeenCalledWith(180)
  })

  it('keeps nothing for a press that did not move; a double-click returns the own size', () => {
    const bridge = installBridge()
    render(<Ribbon snapshot={snapshot({ scale: 150 })} onAddClock={vi.fn()} refused={vi.fn()} />)
    const handle = grip()
    fireEvent.pointerDown(handle, { button: 0, screenX: 0, screenY: 0 })
    fireEvent.pointerUp(handle)
    expect(bridge.SetScale).not.toHaveBeenCalled()
    fireEvent.doubleClick(handle)
    expect(bridge.SetScale).toHaveBeenCalledWith(percentOfWhole)
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
