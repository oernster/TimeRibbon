import { act, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { about, installBridge } from './fakeBridge'
import { About, Licence } from './Help'
import { autoScroll } from './autoScroll'

afterEach(() => vi.useRealTimers())

async function open(panel: 'about' | 'licence') {
  const bridge = installBridge()
  const onClose = vi.fn()
  await act(async () => {
    render(panel === 'about' ? <About onClose={onClose} /> : <Licence onClose={onClose} />)
  })
  return { bridge, onClose }
}

describe('About (FR-607)', () => {
  it('shows the icon, the name and version, the author, the copyright, then every credit, in that order', async () => {
    await open('about')
    const body = document.querySelector('.panel-body') as HTMLElement
    const order = ['IMG', 'Product 0.1.0', 'by The Author', '© The Author', 'Credits'].map((mark) =>
      Array.from(body.querySelectorAll('*')).findIndex((el) => (mark === 'IMG' ? el.tagName === mark : el.textContent === mark)),
    )
    expect(order.every((at) => at >= 0)).toBe(true)
    expect([...order].sort((a, b) => a - b)).toEqual(order)
    const credits = Array.from(body.querySelectorAll('.credits li')).map((li) => li.textContent)
    expect(credits).toEqual(about.credits.map((c) => `${c.name}, ${c.licence}: ${c.role}`))
  })

  it('opens on Close; Close and Escape both return to the ribbon', async () => {
    const { onClose } = await open('about')
    expect(document.activeElement?.textContent).toBe('Close')
    fireEvent.click(screen.getByText('Close'))
    fireEvent.keyDown(screen.getByText('About'), { key: 'Escape' })
    expect(onClose).toHaveBeenCalledTimes(2)
  })

  it('says why when About cannot be read', async () => {
    const bridge = installBridge()
    bridge.About.mockRejectedValueOnce('no facade')
    await act(async () => {
      render(<About onClose={vi.fn()} />)
    })
    expect(screen.getByRole('alert').textContent).toBe('no facade')
  })
})

describe('Licence (FR-608)', () => {
  it('shows the whole of the terms the facade answers', async () => {
    await open('licence')
    expect(document.querySelector('.licence-text')?.textContent).toBe('GNU GENERAL PUBLIC LICENSE\nVersion 3')
  })

  it('reads itself when it holds more than fits (FR-609)', async () => {
    vi.useFakeTimers()
    await open('licence')
    const body = document.querySelector('.panel-body') as HTMLElement
    let top = 0
    Object.defineProperty(body, 'scrollTop', { get: () => top, set: (value: number) => (top = value) })
    Object.defineProperty(body, 'scrollHeight', { get: () => 1000 })
    Object.defineProperty(body, 'clientHeight', { get: () => 0 })
    act(() => vi.advanceTimersByTime(autoScroll.START_HOLD_MS - autoScroll.TICK_MS))
    expect(top).toBe(0)
    act(() => vi.advanceTimersByTime(autoScroll.TICK_MS * 20))
    expect(top).toBeGreaterThan(0)
  })
})
