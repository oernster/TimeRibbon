import { describe, expect, it, vi } from 'vitest'
import { backgroundReporter, rgbOf } from './background'

describe('rgbOf', () => {
  it('reads an opaque computed colour', () => {
    expect(rgbOf('rgb(7, 36, 49)')).toEqual([7, 36, 49])
    expect(rgbOf('rgba(7, 36, 49, 1)')).toEqual([7, 36, 49])
  })

  it('refuses a colour that is not opaque or not rgb at all', () => {
    expect(rgbOf('rgba(0, 0, 0, 0)')).toBeNull()
    expect(rgbOf('transparent')).toBeNull()
    expect(rgbOf('')).toBeNull()
  })
})

describe('backgroundReporter', () => {
  it('reports the background once per change; also when the system switches light or dark', () => {
    const report = vi.fn()
    const listeners: Array<() => void> = []
    const view = Object.create(window) as Window
    Object.defineProperty(view, 'matchMedia', {
      value: () => ({ addEventListener: (_: string, heard: () => void) => listeners.push(heard), removeEventListener: vi.fn() }),
    })
    const reporter = backgroundReporter(report, view)
    document.body.style.backgroundColor = 'rgb(7, 36, 49)'
    reporter.check()
    reporter.check()
    expect(report.mock.calls).toEqual([[7, 36, 49]])
    document.body.style.backgroundColor = 'rgb(246, 247, 249)'
    listeners.forEach((heard) => heard())
    expect(report.mock.calls).toEqual([[7, 36, 49], [246, 247, 249]])
    reporter.stop()
  })
})
