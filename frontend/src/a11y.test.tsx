import { render, waitFor } from '@testing-library/react'
import type { ReactElement } from 'react'
import { describe, expect, it, vi } from 'vitest'
import { installBridge, snapshot } from './fakeBridge'
import { About, Licence, Update } from './Help'
import { Settings } from './Settings'
import { Surface } from './Surface'
import type { SunMap } from './wire'

/** Everything a person can operate: buttons, links and the grip. */
const controls = 'button, a[href], [role="button"], [role="separator"]'

/** A control whose visible text holds no letter or digit shows only an icon, a picture or a glyph. */
const words = /[\p{L}\p{N}]/u

/** iconOnlyWithoutNameOrTip answers each icon-only control in root lacking an accessible name or a tooltip. */
function iconOnlyWithoutNameOrTip(root: HTMLElement): string[] {
  return [...root.querySelectorAll<HTMLElement>(controls)]
    .filter((control) => !words.test(control.textContent ?? ''))
    .filter((control) => !control.getAttribute('aria-label')?.trim() || !control.getAttribute('title')?.trim())
    .map((control) => control.outerHTML)
}

/** A sun map that is on, pulled out below a horizontal ribbon, so the handle and the map both show. */
const pulledOut: SunMap = {
  on: true, pullOut: true, side: 'bottom', shown: true, ribbon: { x: 72, y: 0, width: 336, height: 106 },
  map: { x: 0, y: 106, width: 480, height: 240 }, latitude: 0, longitude: 0, marks: [{ label: 'Mum', latitude: 51.5, longitude: -0.1 }],
}

// Each surface the window can show, drawn as a person first meets it.
const surfaces: Record<string, () => ReactElement> = {
  'the ribbon with its sun map': () => <Surface snapshot={snapshot({ sunMap: pulledOut })} onAddClock={vi.fn()} refused={vi.fn()} />,
  'the empty ribbon': () => <Surface snapshot={snapshot({ cells: [] })} onAddClock={vi.fn()} refused={vi.fn()} />,
  'Settings': () => <Settings snapshot={snapshot()} startAdding={false} reload={vi.fn()} onClose={vi.fn()} />,
  'Settings on the place search': () => <Settings snapshot={snapshot()} startAdding reload={vi.fn()} onClose={vi.fn()} />,
  'About': () => <About onClose={vi.fn()} />,
  'Licence': () => <Licence onClose={vi.fn()} />,
  'the update panel': () => <Update status={{ current: '2.0.0', latest: 'v2.1.0', updateAvailable: true }} onClose={vi.fn()} />,
}

describe('icon-only controls (NFR-U-4)', () => {
  for (const [name, draw] of Object.entries(surfaces)) {
    it(`each on ${name} carries an accessible name and a tooltip`, async () => {
      installBridge()
      const { container } = render(draw())
      // Panels fill in once the bridge answers; the check is of what a person then meets.
      await waitFor(() => expect(container.querySelector(controls)).not.toBeNull())
      expect(iconOnlyWithoutNameOrTip(container)).toEqual([])
    })
  }

  it('finds an icon-only control that lacks either, so it can catch one', () => {
    const { container } = render(
      <div>
        <button type="button" aria-label="Close">×</button>
        <button type="button" title="Close">×</button>
        <button type="button">Close</button>
      </div>,
    )
    expect(iconOnlyWithoutNameOrTip(container)).toHaveLength(2)
  })
})
