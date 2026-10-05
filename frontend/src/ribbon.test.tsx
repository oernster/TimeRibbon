import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { api } from './api'
import { addClockTip } from './ArtButton'
import { cell, installBridge, snapshot } from './fakeBridge'
import { Ribbon } from './Ribbon'

describe('Ribbon', () => {
  it('shows each clock in order with its own date (FR-102, FR-202)', () => {
    installBridge()
    render(<Ribbon snapshot={snapshot()} onAddClock={vi.fn()} refused={vi.fn()} />)
    const groups = screen.getAllByRole('group')
    expect(groups.map((group) => group.getAttribute('aria-label'))).toEqual([
      'New York, 16:37, Sunday, 27 September',
      'Sydney, 06:37, Monday, 28 September',
    ])
  })

  it('shows a clock that cannot be read in words, with no time (FR-706, NFR-U-2)', () => {
    installBridge()
    const broken = cell({ id: 'b', label: 'Gran', zoneMark: '', time: '', date: '', problem: 'Unknown time zone: Not/AZone' })
    render(<Ribbon snapshot={snapshot({ cells: [broken] })} onAddClock={vi.fn()} refused={vi.fn()} />)
    expect(screen.getByText('Unknown time zone: Not/AZone')).toBeTruthy()
    expect(screen.queryByText(/\d\d:\d\d/)).toBeNull()
  })

  it('draws cells at the size the snapshot names, marked for its text sizes (FR-610)', () => {
    installBridge()
    const small = { digital: { width: 144, height: 72 }, analogue: { width: 144, height: 124 }, prompt: { width: 176, height: 184 }, padding: 6, handleLane: 16 }
    render(<Ribbon snapshot={snapshot({ size: 'small', layout: small })} onAddClock={vi.fn()} refused={vi.fn()} />)
    const ribbon = screen.getAllByRole('group')[0].closest('.ribbon') as HTMLElement
    expect(ribbon.classList.contains('small')).toBe(true)
    expect(ribbon.style.getPropertyValue('--cell-w')).toBe('144px')
    expect(ribbon.style.getPropertyValue('--cell-h')).toBe('72px')
    // With the sun map off there is no handle, so no lane (FR-903).
    expect([...ribbon.classList].some((name) => name.startsWith('lane-'))).toBe(false)
  })

  it('offers Add clock on an empty ribbon (FR-107)', () => {
    installBridge()
    const onAddClock = vi.fn()
    render(<Ribbon snapshot={snapshot({ cells: [] })} onAddClock={onAddClock} refused={vi.fn()} />)
    fireEvent.click(screen.getByLabelText(addClockTip))
    expect(onAddClock).toHaveBeenCalled()
  })

  // The tab, the drag, the wheel and the drawn report are ribbonkit's band (Band.test.tsx); this
  // only checks the clocks hand it their calls.
  it('opens the native menu on right-click (FR-108)', () => {
    const bridge = installBridge()
    render(<Ribbon snapshot={snapshot()} onAddClock={vi.fn()} refused={vi.fn()} />)
    fireEvent.contextMenu(screen.getAllByRole('group')[0])
    expect(bridge.ShowContextMenu).toHaveBeenCalled()
  })
})

describe('api', () => {
  it('answers null and says why when Go is not there or refuses', async () => {
    delete window.go
    const refused = vi.fn()
    expect(await api.snapshot(refused)).toBeNull()
    expect(refused).toHaveBeenCalledWith('TimeRibbon is not running behind this page')
    const bridge = installBridge()
    bridge.AddClock.mockRejectedValueOnce('unknown time zone: X')
    expect(await api.addClock('X', refused)).toBeNull()
    expect(refused).toHaveBeenLastCalledWith('unknown time zone: X')
  })
})
