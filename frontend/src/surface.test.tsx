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

// The handle, its arrows, its lane and where the parts stand are ribbonkit's (PullOut.test.tsx,
// Band.test.tsx); these check the sun map is what is pulled out, in TimeRibbon's words.
describe('Surface (FR-902, FR-903, FR-910)', () => {
  it('gives a vertical ribbon a handle that asks Go to pull the map out, then to put it back', () => {
    const bridge = installBridge()
    const { rerender } = render(<Surface snapshot={snapshot({ orientation: 'vertical', sunMap: sunMap() })} onAddClock={vi.fn()} refused={vi.fn()} />)
    // The handle stands in a lane of its own on the map's side, as deep as the handle (Amendment 23).
    const ribbon = document.querySelector('.ribbon') as HTMLElement
    expect(ribbon.classList.contains('lane-left')).toBe(true)
    expect(ribbon.style.getPropertyValue('--lane')).toBe('16px')
    expect((document.querySelector('.surface') as HTMLElement).style.getPropertyValue('--handle-width')).toBe('16px')
    fireEvent.click(screen.getByRole('button', { name: openPullOut }))
    expect(bridge.TogglePullOut).toHaveBeenCalledOnce()
    rerender(<Surface snapshot={snapshot({ orientation: 'vertical', sunMap: sunMap({ pullOut: true }) })} onAddClock={vi.fn()} refused={vi.fn()} />)
    expect(screen.getByRole('button', { name: closePullOut }).getAttribute('title')).toBe(closePullOut)
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

  it('shows no map with the tab', () => {
    installBridge()
    render(<Surface snapshot={snapshot({ collapsed: true, sunMap: sunMap({ side: 'bottom', shown: true, marks: [{ label: 'Mum', latitude: 0, longitude: 0 }] }) })} onAddClock={vi.fn()} refused={vi.fn()} />)
    expect(screen.queryByText('Mum')).toBeNull()
    expect(screen.getByTestId('tab')).toBeTruthy()
  })
})
