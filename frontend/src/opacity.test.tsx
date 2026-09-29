import { fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { installBridge, snapshot } from './fakeBridge'
import { opacityProperty, percentOfWhole, showOpacity } from './opacity'
import { OpacitySlider } from './OpacitySlider'

afterEach(() => document.documentElement.style.removeProperty(opacityProperty))

describe('opacity (FR-622)', () => {
  it('draws the window at the fraction of the percentage', () => {
    showOpacity(40)
    expect(document.documentElement.style.getPropertyValue(opacityProperty)).toBe('0.4')
  })

  it('offers the least the setting allows up to wholly opaque, showing the chosen value', () => {
    installBridge()
    render(<OpacitySlider snapshot={snapshot({ opacity: 70, minOpacity: 20 })} refused={vi.fn()} then={vi.fn()} />)
    const slider = screen.getByRole('slider') as HTMLInputElement
    expect([slider.min, slider.max, slider.value]).toEqual(['20', String(percentOfWhole), '70'])
    expect(screen.getByText('70%')).toBeTruthy()
  })

  it('follows the slider while it moves and keeps the choice once, when it is let go', async () => {
    const bridge = installBridge()
    const then = vi.fn()
    render(<OpacitySlider snapshot={snapshot()} refused={vi.fn()} then={then} />)
    const slider = screen.getByRole('slider')
    fireEvent.change(slider, { target: { value: '55' } })
    expect(document.documentElement.style.getPropertyValue(opacityProperty)).toBe('0.55')
    expect(bridge.SetOpacity).not.toHaveBeenCalled()
    fireEvent.pointerUp(slider)
    await vi.waitFor(() => expect(then).toHaveBeenCalledOnce())
    expect(bridge.SetOpacity).toHaveBeenCalledOnce()
    expect(bridge.SetOpacity).toHaveBeenCalledWith(55)
  })

  it('keeps nothing when the slider was not moved', () => {
    const bridge = installBridge()
    render(<OpacitySlider snapshot={snapshot()} refused={vi.fn()} then={vi.fn()} />)
    fireEvent.keyUp(screen.getByRole('slider'))
    expect(bridge.SetOpacity).not.toHaveBeenCalled()
  })
})
