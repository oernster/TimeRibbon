import { renderHook, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { installBridge, snapshot } from './fakeBridge'
import { cellWidthNeeded, useMeasuredCells } from './measure'

// jsdom lays nothing out, so a box that shrinks to its widest line is given that line's length in
// characters times this; every other box has no width at all.
const perCharacter = 7

beforeEach(() => {
  vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockImplementation(function (this: HTMLElement) {
    const widest = this.style.width === 'max-content' ? Math.max(...[...this.children].map((line) => line.textContent?.length ?? 0)) : 0
    return { width: widest * perCharacter } as DOMRect
  })
})

afterEach(() => {
  vi.restoreAllMocks()
  document.head.querySelectorAll('style').forEach((style) => style.remove())
})

describe('cellWidthNeeded (FR-620)', () => {
  const samples = { times: ['12:59 PM'], dates: ['Tue 1/5'] }

  it('takes the widest time and date for a digital cell, the widest date alone for an analogue one', () => {
    expect(cellWidthNeeded(samples, 'large', false)).toBe('12:59 PM'.length * perCharacter)
    expect(cellWidthNeeded(samples, 'large', true)).toBe('Tue 1/5'.length * perCharacter)
  })

  it('adds the padding and the divider a cell after the first carries, leaving nothing behind', () => {
    const style = document.createElement('style')
    style.textContent = '.cell { padding-left: 12px; padding-right: 12px } .cell + .cell { border-left: 1px solid black }'
    document.head.appendChild(style)
    expect(cellWidthNeeded(samples, 'small', true)).toBe('Tue 1/5'.length * perCharacter + 12 + 12 + 1)
    expect(document.querySelector('.ribbon')).toBeNull()
  })
})

describe('useMeasuredCells (FR-620)', () => {
  it('reports the measurement for the choices it was taken under, once per set of choices, then takes the snapshot again', async () => {
    const bridge = installBridge()
    const load = vi.fn()
    const refused = vi.fn()
    const shown = snapshot({ size: 'small', format: '12h', dateFormat: 'dmy' })
    const { rerender } = renderHook(({ current }) => useMeasuredCells(current, load, refused), { initialProps: { current: shown } })
    await waitFor(() => expect(load).toHaveBeenCalledOnce())
    expect(bridge.SetMeasured).toHaveBeenCalledWith({
      size: 'small', style: 'digital', format: '12h', dateFormat: 'dmy', cellWidth: 'Wednesday, 30 September'.length * perCharacter,
    })
    rerender({ current: { ...shown, cells: [] } })
    await Promise.resolve()
    expect(bridge.SetMeasured).toHaveBeenCalledOnce()
    expect(refused).not.toHaveBeenCalled()
  })

  it('measures nothing before the first snapshot', async () => {
    const bridge = installBridge()
    renderHook(() => useMeasuredCells(null, vi.fn(), vi.fn()))
    await Promise.resolve()
    expect(bridge.TextSamples).not.toHaveBeenCalled()
  })
})
