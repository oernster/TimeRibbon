// A stand-in for the Go facade in tests: every call is recorded, every answer is canned.

import { vi } from 'vitest'
import type { About, Cell, Place, Snapshot } from './wire'

export const about: About = {
  name: 'Product', version: '0.1.0', author: 'The Author', copyright: '© The Author',
  credits: [
    { name: 'Go standard library', licence: 'BSD-3-Clause', role: 'the language and its runtime' },
    { name: 'Wails v2', licence: 'MIT', role: 'the desktop shell' },
  ],
}

export function cell(overrides: Partial<Cell> = {}): Cell {
  return {
    id: 'ny', label: 'New York', zone: 'America/New_York', zoneMark: 'EDT', time: '16:37',
    date: 'Sunday, 27 September', hourAngle: 138.5, minuteAngle: 222, problem: '', ...overrides,
  }
}

export function snapshot(overrides: Partial<Snapshot> = {}): Snapshot {
  return {
    cells: [cell(), cell({ id: 'syd', label: 'Sydney', zone: 'Australia/Sydney', zoneMark: 'AEST', time: '06:37', date: 'Monday, 28 September' })],
    style: 'digital', size: 'large', colour: 'classic', format: '24h', dateFormat: 'day-month', orientation: 'horizontal', theme: 'system', alwaysOnTop: false,
    layout: { digital: { width: 176, height: 92 }, analogue: { width: 176, height: 176 }, prompt: { width: 176, height: 184 }, padding: 6 },
    refreshInMs: 60000, notices: [], scrolls: false, dragThreshold: { width: 4, height: 4 }, startLabel: 'Start at sign-in',
    collapsed: false,
    ...overrides,
  }
}

export const places: Place[] = [
  { zone: 'Asia/Kolkata', label: 'Kolkata', country: 'India' },
  { zone: 'Europe/Oslo', label: 'Oslo', country: 'Norway' },
]

/** installBridge puts a recording facade on window and answers it. */
export function installBridge() {
  const bridge = {
    Snapshot: vi.fn(async () => snapshot()),
    AddClock: vi.fn(async () => 'new'),
    RenameClock: vi.fn(async () => undefined),
    RezoneClock: vi.fn(async () => undefined),
    RemoveClock: vi.fn(async () => undefined),
    SearchPlaces: vi.fn(async () => places),
    SetSize: vi.fn(async () => undefined),
    SetFormat: vi.fn(async () => undefined),
    SetDateFormat: vi.fn(async () => undefined),
    SetTheme: vi.fn(async () => undefined),
    SetAlwaysOnTop: vi.fn(async () => undefined),
    StartWithWindows: vi.fn(async () => false),
    SetStartWithWindows: vi.fn(async () => undefined),
    DismissNotices: vi.fn(async () => undefined),
    SetScrollbar: vi.fn(async () => undefined),
    SetPixelRatio: vi.fn(async () => undefined),
    SetBackground: vi.fn(async () => undefined),
    RibbonDrawn: vi.fn(async () => undefined),
    ShowContextMenu: vi.fn(async () => undefined),
    OpenPanel: vi.fn(async () => undefined),
    ClosePanel: vi.fn(async () => undefined),
    Hide: vi.fn(async () => undefined),
    OpenDonation: vi.fn(async () => undefined),
    OpenUpdate: vi.fn(async () => undefined),
    SkipUpdate: vi.fn(async () => undefined),
    About: vi.fn(async () => about),
    Licence: vi.fn(async () => 'GNU GENERAL PUBLIC LICENSE\nVersion 3'),
  }
  window.go = { main: { App: bridge } }
  window.WailsInvoke = vi.fn()
  return bridge
}
