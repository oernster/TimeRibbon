import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { installBridge, snapshot } from './fakeBridge'
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

  it('gives a horizontal ribbon no handle and draws its map with each clock marked by name', () => {
    const shown = sunMap({
      side: 'bottom', shown: true, ribbon: { x: 72, y: 0, width: 336, height: 106 }, map: { x: 0, y: 106, width: 480, height: 240 },
      marks: [{ label: 'Mum', latitude: 51.5, longitude: -0.1 }],
    })
    const bridge = installBridge()
    render(<Surface snapshot={snapshot({ sunMap: shown })} onAddClock={vi.fn()} refused={vi.fn()} />)
    expect(screen.queryByRole('button', { name: openPullOut })).toBeNull()
    expect(screen.getByText('Mum')).toBeTruthy()
    // A right-click on the map offers the ribbon's own menu (FR-108).
    fireEvent.contextMenu(screen.getByText('Mum').closest('.sun-map') as HTMLElement)
    expect(bridge.ShowContextMenu).toHaveBeenCalledOnce()
  })

  it('shows no map with the tab', () => {
    installBridge()
    render(<Surface snapshot={snapshot({ collapsed: true, sunMap: sunMap({ side: 'bottom', shown: true, marks: [{ label: 'Mum', latitude: 0, longitude: 0 }] }) })} onAddClock={vi.fn()} refused={vi.fn()} />)
    expect(screen.queryByText('Mum')).toBeNull()
    expect(screen.getByTestId('tab')).toBeTruthy()
  })
})
