import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { api } from './api'
import { addClockTip } from './ArtButton'
import { cell, installBridge, snapshot } from './fakeBridge'
import { scrollbarThickness } from './scrollbar'
import { Strip } from './Strip'

describe('Strip', () => {
  it('shows each clock in order with its own date (FR-102, FR-202)', () => {
    installBridge()
    render(<Strip snapshot={snapshot()} onAddClock={vi.fn()} refused={vi.fn()} />)
    const groups = screen.getAllByRole('group')
    expect(groups.map((group) => group.getAttribute('aria-label'))).toEqual([
      'New York, 16:37, Sunday, 27 September',
      'Sydney, 06:37, Monday, 28 September',
    ])
  })

  it('shows a clock that cannot be read in words, with no time (FR-706, NFR-U-2)', () => {
    installBridge()
    const broken = cell({ id: 'b', label: 'Gran', zoneMark: '', time: '', date: '', problem: 'Unknown time zone: Not/AZone' })
    render(<Strip snapshot={snapshot({ cells: [broken] })} onAddClock={vi.fn()} refused={vi.fn()} />)
    expect(screen.getByText('Unknown time zone: Not/AZone')).toBeTruthy()
    expect(screen.queryByText(/\d\d:\d\d/)).toBeNull()
  })

  it('draws cells at the size the snapshot names, marked for its text sizes (FR-610)', () => {
    installBridge()
    const small = { digital: { width: 144, height: 72 }, analogue: { width: 144, height: 124 }, prompt: { width: 176, height: 184 }, padding: 6 }
    render(<Strip snapshot={snapshot({ size: 'small', layout: small })} onAddClock={vi.fn()} refused={vi.fn()} />)
    const strip = screen.getAllByRole('group')[0].closest('.strip') as HTMLElement
    expect(strip.classList.contains('small')).toBe(true)
    expect(strip.style.getPropertyValue('--cell-w')).toBe('144px')
    expect(strip.style.getPropertyValue('--cell-h')).toBe('72px')
  })

  it('offers Add clock on an empty strip (FR-107)', () => {
    installBridge()
    const onAddClock = vi.fn()
    render(<Strip snapshot={snapshot({ cells: [] })} onAddClock={onAddClock} refused={vi.fn()} />)
    fireEvent.click(screen.getByLabelText(addClockTip))
    expect(onAddClock).toHaveBeenCalled()
  })

  it('starts a drag only past the threshold and never from a control (FR-401, FR-402)', () => {
    installBridge()
    render(<Strip snapshot={snapshot({ cells: [], notices: ['kept aside'] })} onAddClock={vi.fn()} refused={vi.fn()} />)
    const strip = screen.getByText('No clocks yet').closest('.strip') as HTMLElement
    fireEvent.pointerDown(strip, { button: 0, screenX: 100, screenY: 100 })
    fireEvent.pointerMove(strip, { buttons: 1, screenX: 103, screenY: 102 })
    expect(window.WailsInvoke).not.toHaveBeenCalled()
    fireEvent.pointerMove(strip, { buttons: 1, screenX: 105, screenY: 100 })
    expect(window.WailsInvoke).toHaveBeenCalledWith('drag')
    const invoke = window.WailsInvoke as ReturnType<typeof vi.fn>
    invoke.mockClear()
    const button = screen.getByLabelText(addClockTip)
    fireEvent.pointerDown(button, { button: 0, screenX: 100, screenY: 100 })
    fireEvent.pointerMove(button, { buttons: 1, screenX: 140, screenY: 100 })
    expect(invoke).not.toHaveBeenCalled()
  })

  it('moves only a scrolling horizontal strip along with a plain wheel (FR-106)', () => {
    installBridge()
    const along = (overrides: Parameters<typeof snapshot>[0], deltaX: number) => {
      const view = render(<Strip snapshot={snapshot(overrides)} onAddClock={vi.fn()} refused={vi.fn()} />)
      const strip = view.container.querySelector('.strip') as HTMLElement
      let left = 0
      Object.defineProperty(strip, 'scrollLeft', { get: () => left, set: (value: number) => (left = value) })
      fireEvent.wheel(strip, { deltaX, deltaY: 100 })
      view.unmount()
      return left
    }
    expect(along({ orientation: 'horizontal', scrolls: true }, 0)).toBe(100)
    expect(along({ orientation: 'horizontal', scrolls: true }, 30)).toBe(0)
    expect(along({ orientation: 'horizontal', scrolls: false }, 0)).toBe(0)
    expect(along({ orientation: 'vertical', scrolls: true }, 0)).toBe(0)
  })

  it('measures the scroll bar as the room it takes from a box that must scroll (FR-106)', () => {
    expect(scrollbarThickness()).toBe(0)
    const offset = vi.spyOn(HTMLElement.prototype, 'offsetHeight', 'get').mockReturnValue(100)
    const client = vi.spyOn(HTMLElement.prototype, 'clientHeight', 'get').mockReturnValue(85)
    try {
      expect(scrollbarThickness()).toBe(15)
      expect(document.body.children.length).toBe(0)
    } finally {
      offset.mockRestore()
      client.mockRestore()
    }
  })

  it('opens the native menu on right-click (FR-108)', () => {
    const bridge = installBridge()
    render(<Strip snapshot={snapshot()} onAddClock={vi.fn()} refused={vi.fn()} />)
    fireEvent.contextMenu(screen.getAllByRole('group')[0])
    expect(bridge.ShowContextMenu).toHaveBeenCalled()
  })
})

describe('api', () => {
  it('answers null and says why when Go is not there or refuses', async () => {
    delete window.go
    const refused = vi.fn()
    expect(await api.snapshot(refused)).toBeNull()
    expect(refused).toHaveBeenCalledWith('TimeStrip is not running behind this page')
    const bridge = installBridge()
    bridge.AddClock.mockRejectedValueOnce('unknown time zone: X')
    expect(await api.addClock('X', refused)).toBeNull()
    expect(refused).toHaveBeenLastCalledWith('unknown time zone: X')
  })
})
