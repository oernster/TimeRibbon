import { fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { installBridge, snapshot } from './fakeBridge'
import { percentOfWhole } from '@oernster/ribbonkit'
import { Ribbon } from './Ribbon'
import { scaleGripTip } from './ScaleGrip'

// The ribbon the grip sits in, in the page's pixels across it.
const thickness = 100

afterEach(() => vi.restoreAllMocks())

describe('ScaleGrip (FR-623)', () => {
  function grip() {
    const element = screen.getByRole('separator', { name: scaleGripTip })
    vi.spyOn(element.parentElement as HTMLElement, 'getBoundingClientRect').mockReturnValue({ width: thickness * 2, height: thickness } as DOMRect)
    return element
  }

  it('draws the clocks at the snapshot scale and hands Go the drag: its start, each move and its end', async () => {
    const bridge = installBridge()
    render(<Ribbon snapshot={snapshot({ scale: 120 })} onAddClock={vi.fn()} refused={vi.fn()} />)
    const ribbon = document.querySelector('.ribbon') as HTMLElement
    expect(ribbon.style.zoom).toBe('1.2')
    const handle = grip()
    fireEvent.pointerDown(handle, { button: 0, screenX: 5, screenY: 7 })
    expect(bridge.BeginScale).toHaveBeenCalledWith(thickness, 5, 7)
    fireEvent.pointerMove(handle, { screenX: 5, screenY: 57 })
    await vi.waitFor(() => expect(bridge.DragScale).toHaveBeenCalledWith(5, 57))
    expect(bridge.EndScale).not.toHaveBeenCalled()
    fireEvent.pointerUp(handle, { screenX: 5, screenY: 60 })
    expect(bridge.EndScale).toHaveBeenCalledWith(5, 60)
  })

  it('measures a vertical ribbon across its width', () => {
    const bridge = installBridge()
    render(<Ribbon snapshot={snapshot({ orientation: 'vertical' })} onAddClock={vi.fn()} refused={vi.fn()} />)
    fireEvent.pointerDown(grip(), { button: 0, screenX: 0, screenY: 0 })
    expect(bridge.BeginScale).toHaveBeenCalledWith(thickness * 2, 0, 0)
  })

  it('sends one move at a time, the newest pointer following it', async () => {
    const bridge = installBridge()
    let finish: () => void = () => undefined
    bridge.DragScale.mockImplementationOnce(() => new Promise<undefined>((done) => { finish = () => done(undefined) }))
    render(<Ribbon snapshot={snapshot()} onAddClock={vi.fn()} refused={vi.fn()} />)
    const handle = grip()
    fireEvent.pointerDown(handle, { button: 0, screenX: 0, screenY: 0 })
    fireEvent.pointerMove(handle, { screenX: 0, screenY: 10 })
    fireEvent.pointerMove(handle, { screenX: 0, screenY: 20 })
    fireEvent.pointerMove(handle, { screenX: 0, screenY: 30 })
    expect(bridge.DragScale).toHaveBeenCalledTimes(1)
    finish()
    await vi.waitFor(() => expect(bridge.DragScale).toHaveBeenCalledTimes(2))
    expect(bridge.DragScale).toHaveBeenLastCalledWith(0, 30)
  })

  it('a double-click returns the own size; a move with no press sends nothing', () => {
    const bridge = installBridge()
    render(<Ribbon snapshot={snapshot({ scale: 150 })} onAddClock={vi.fn()} refused={vi.fn()} />)
    const handle = grip()
    fireEvent.pointerMove(handle, { screenX: 0, screenY: 30 })
    expect(bridge.DragScale).not.toHaveBeenCalled()
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
