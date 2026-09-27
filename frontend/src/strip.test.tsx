import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { api } from './api'
import { addClockTip } from './ArtButton'
import { cell, installBridge, snapshot } from './fakeBridge'
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
