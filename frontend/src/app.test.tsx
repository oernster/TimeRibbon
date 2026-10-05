import { render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { App } from './App'
import { installBridge } from './fakeBridge'
import { installEvents } from '@oernster/ribbonkit/testing'

afterEach(() => {
  delete window.runtime
})

// The shell's own rules (panels, refresh, theme, opacity) are ribbonkit's (shell.test.tsx); this
// checks the one word TimeRibbon adds to them.
describe('App', () => {
  it('opens Settings on the place search when Go says add-clock (FR-107)', async () => {
    const bridge = installBridge()
    const fire = installEvents()
    render(<App />)
    await waitFor(() => expect(bridge.Snapshot).toHaveBeenCalled())
    await fire('open-panel', 'add-clock')
    expect(bridge.OpenPanel).toHaveBeenCalledWith('settings')
    // Settings is the panel either way; add-clock is what puts the cursor in the place search.
    await waitFor(() => expect(document.activeElement).toBe(screen.getByLabelText('Search places')))
  })
})
