import { render, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { installBridge } from './fakeBridge'
import { usePanelFit } from './panelFit'

// jsdom lays nothing out, so a panel measured with room for its content is this tall per child.
const perChild = 100

function Panel({ refused }: { refused: () => void }) {
  const panel = usePanelFit<HTMLDivElement>(refused)
  return (
    <div ref={panel} data-testid="panel" style={{ height: '50px' }}>
      <p>one</p>
      <p>two</p>
    </div>
  )
}

beforeEach(() => {
  vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockImplementation(function (this: HTMLElement) {
    return { height: this.style.height === 'auto' ? this.children.length * perChild : 0 } as DOMRect
  })
})

afterEach(() => vi.restoreAllMocks())

describe('usePanelFit (FR-621)', () => {
  it('tells Go the height the content needs, again whenever anything inside changes it, never twice the same', async () => {
    const bridge = installBridge()
    const refused = vi.fn()
    const { getByTestId } = render(<Panel refused={refused} />)
    expect(bridge.FitPanel).toHaveBeenLastCalledWith(2 * perChild)
    const panel = getByTestId('panel')
    expect(panel.style.height).toBe('50px')
    panel.appendChild(document.createElement('p'))
    await waitFor(() => expect(bridge.FitPanel).toHaveBeenLastCalledWith(3 * perChild))
    panel.firstElementChild?.setAttribute('class', 'same height')
    await Promise.resolve()
    expect(bridge.FitPanel).toHaveBeenCalledTimes(2)
    expect(refused).not.toHaveBeenCalled()
  })
})
