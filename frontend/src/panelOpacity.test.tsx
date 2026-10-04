import { act, render, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { App } from './App'
import { installBridge, snapshot } from './fakeBridge'
import { opacityProperty } from './opacity'

afterEach(() => {
  document.documentElement.style.removeProperty(opacityProperty)
  delete window.runtime
})

/** installEvents puts a runtime on window whose handlers can be fired by name. */
function installEvents() {
  const handlers = new Map<string, (...data: unknown[]) => void>()
  window.runtime = {
    EventsOn: vi.fn((name: string, callback: (...data: unknown[]) => void) => {
      handlers.set(name, callback)
      return () => handlers.delete(name)
    }),
  } as unknown as typeof window.runtime
  return (name: string, ...data: unknown[]) => act(() => handlers.get(name)?.(...data))
}

describe('the opacity is the ribbon\'s alone (FR-622)', () => {
  it('draws the ribbon at the chosen opacity and a panel wholly opaque', async () => {
    const bridge = installBridge()
    bridge.Snapshot.mockImplementation(async () => snapshot({ opacity: 40 }))
    const fire = installEvents()
    render(<App />)
    const shown = () => document.documentElement.style.getPropertyValue(opacityProperty)
    await waitFor(() => expect(shown()).toBe('0.4'))
    await fire('open-panel', 'settings')
    await waitFor(() => expect(shown()).toBe('1'))
  })
})
