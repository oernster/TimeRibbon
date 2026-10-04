import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { installBridge, snapshot } from './fakeBridge'
import { SunMap as SunMapView } from './SunMap'
import { Surface, closePullOut, openPullOut } from './Surface'
import type { SunMap } from './wire'

function sunMap(overrides: Partial<SunMap> = {}): SunMap {
  return {
    on: true, pullOut: false, side: 'left', shown: false,
    ribbon: { x: 0, y: 0, width: 176, height: 400 }, map: { x: 0, y: 0, width: 0, height: 0 },
    latitude: 0, longitude: 0, marks: [], ...overrides,
  }
}

describe('Surface (FR-902, FR-903, FR-910)', () => {
  it('gives a vertical ribbon a handle that asks Go to pull the map out, then to put it back', () => {
    const bridge = installBridge()
    const { rerender } = render(<Surface snapshot={snapshot({ orientation: 'vertical', sunMap: sunMap() })} onAddClock={vi.fn()} refused={vi.fn()} />)
    fireEvent.click(screen.getByRole('button', { name: openPullOut }))
    expect(bridge.TogglePullOut).toHaveBeenCalledOnce()
    rerender(<Surface snapshot={snapshot({ orientation: 'vertical', sunMap: sunMap({ pullOut: true }) })} onAddClock={vi.fn()} refused={vi.fn()} />)
    expect(screen.getByRole('button', { name: closePullOut }).getAttribute('title')).toBe(closePullOut)
  })

  it('gives a horizontal ribbon the same handle, its arrow pointing the way the map will move (Amendment 22)', () => {
    const bridge = installBridge()
    const closed = sunMap({ side: 'bottom', ribbon: { x: 0, y: 0, width: 336, height: 106 } })
    const { rerender } = render(<Surface snapshot={snapshot({ sunMap: closed })} onAddClock={vi.fn()} refused={vi.fn()} />)
    const handle = screen.getByRole('button', { name: openPullOut })
    expect(handle.textContent).toBe('▼')
    expect(handle.className).toBe('pull-out-handle bottom')
    // The handle stands in a lane of its own on that side, as deep as the handle (Amendment 23).
    const ribbon = document.querySelector('.ribbon') as HTMLElement
    expect(ribbon.classList.contains('lane-bottom')).toBe(true)
    expect(ribbon.style.getPropertyValue('--lane')).toBe('16px')
    expect((handle.closest('.surface') as HTMLElement).style.getPropertyValue('--handle-width')).toBe('16px')
    fireEvent.click(handle)
    expect(bridge.TogglePullOut).toHaveBeenCalledOnce()
    rerender(<Surface snapshot={snapshot({ sunMap: sunMap({ side: 'top', pullOut: true, ribbon: closed.ribbon }) })} onAddClock={vi.fn()} refused={vi.fn()} />)
    expect(screen.getByRole('button', { name: closePullOut }).textContent).toBe('▼')
  })

  it('draws a horizontal ribbon\'s map with each clock marked by name', () => {
    const shown = sunMap({
      side: 'bottom', shown: true, pullOut: true, ribbon: { x: 72, y: 0, width: 336, height: 106 }, map: { x: 0, y: 106, width: 480, height: 240 },
      marks: [{ label: 'Mum', latitude: 51.5, longitude: -0.1 }],
    })
    const bridge = installBridge()
    render(<Surface snapshot={snapshot({ sunMap: shown })} onAddClock={vi.fn()} refused={vi.fn()} />)
    expect(screen.getByText('Mum')).toBeTruthy()
    // A right-click on the map offers the ribbon's own menu (FR-108).
    fireEvent.contextMenu(screen.getByText('Mum').closest('.sun-map') as HTMLElement)
    expect(bridge.ShowContextMenu).toHaveBeenCalledOnce()
  })

  it('answers a right-click and a drag beside a ribbon shorter than its map as the ribbon does, once', () => {
    const shown = sunMap({
      side: 'bottom', shown: true, pullOut: true, ribbon: { x: 72, y: 0, width: 336, height: 106 }, map: { x: 0, y: 106, width: 480, height: 240 },
    })
    const bridge = installBridge()
    window.WailsInvoke = vi.fn()
    render(<Surface snapshot={snapshot({ sunMap: shown })} onAddClock={vi.fn()} refused={vi.fn()} />)
    const surface = document.querySelector('.surface') as HTMLElement
    fireEvent.contextMenu(surface)
    expect(bridge.ShowContextMenu).toHaveBeenCalledOnce()
    // A right-click on the ribbon reaches the surface too; it still shows one menu, not two.
    fireEvent.contextMenu(document.querySelector('.ribbon') as HTMLElement)
    expect(bridge.ShowContextMenu).toHaveBeenCalledTimes(2)
    fireEvent.pointerDown(surface, { button: 0, screenX: 100, screenY: 100 })
    fireEvent.pointerMove(surface, { buttons: 1, screenX: 140, screenY: 100 })
    expect(window.WailsInvoke).toHaveBeenCalledOnce()
    expect(window.WailsInvoke).toHaveBeenCalledWith('drag')
  })

  it('measures each label, then stands London clear of the Berlin dot (FR-914)', () => {
    installBridge()
    // jsdom lays nothing out, so every label is given the size a 12px name takes on screen.
    const width = vi.spyOn(HTMLElement.prototype, 'offsetWidth', 'get').mockReturnValue(40)
    const height = vi.spyOn(HTMLElement.prototype, 'offsetHeight', 'get').mockReturnValue(16)
    try {
      const marks = [{ label: 'London', latitude: 51.51, longitude: -0.13 }, { label: 'Berlin', latitude: 52.52, longitude: 13.4 }]
      render(<SunMapView sunMap={sunMap({ shown: true, marks })} width={708} height={354} dragThreshold={{ width: 4, height: 4 }} refused={vi.fn()} />)
      expect(screen.getByText('London').className).toBe('mark-label left')
      expect(screen.getByText('Berlin').className).toBe('mark-label right')
    } finally {
      width.mockRestore()
      height.mockRestore()
    }
  })

  it('draws the ribbon at the box Go gives, whatever the window\'s width, so one drawn inside the tab fits once the window grows (FR-615)', () => {
    installBridge()
    const width = Object.getOwnPropertyDescriptor(window, 'innerWidth')
    // The page draws an opening ribbon while the window is still the tab, 8 pixels wide; measured
    // 2026-09-29, a vertical ribbon with the sun map on then grew to 175 pixels drawn 8 wide.
    Object.defineProperty(window, 'innerWidth', { configurable: true, value: 8 })
    try {
      render(<Surface snapshot={snapshot({ orientation: 'vertical', sunMap: sunMap({ side: 'right', ribbon: { x: 0, y: 0, width: 140, height: 659.2 } }) })} onAddClock={vi.fn()} refused={vi.fn()} />)
      const part = (document.querySelector('.ribbon') as HTMLElement).closest('.surface-part') as HTMLElement
      expect(part.style.width).toBe('140px')
      expect(part.style.height).toBe('659.2px')
    } finally {
      Object.defineProperty(window, 'innerWidth', width ?? { configurable: true, value: 1024 })
    }
  })

  it('shows no map with the tab', () => {
    installBridge()
    render(<Surface snapshot={snapshot({ collapsed: true, sunMap: sunMap({ side: 'bottom', shown: true, marks: [{ label: 'Mum', latitude: 0, longitude: 0 }] }) })} onAddClock={vi.fn()} refused={vi.fn()} />)
    expect(screen.queryByText('Mum')).toBeNull()
    expect(screen.getByTestId('tab')).toBeTruthy()
  })
})
