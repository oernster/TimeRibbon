// A stand-in for the Go facade in tests: every call is recorded, every answer is canned.

import { vi } from 'vitest'
import type { Cell, Place, Snapshot } from './wire'

export function cell(overrides: Partial<Cell> = {}): Cell {
  return {
    id: 'ny', label: 'New York', zone: 'America/New_York', zoneMark: 'EDT', time: '16:37',
    date: 'Sunday, 27 September', hourAngle: 138.5, minuteAngle: 222, problem: '', ...overrides,
  }
}

export function snapshot(overrides: Partial<Snapshot> = {}): Snapshot {
  return {
    cells: [cell(), cell({ id: 'syd', label: 'Sydney', zone: 'Australia/Sydney', zoneMark: 'AEST', time: '06:37', date: 'Monday, 28 September' })],
    style: 'digital', format: '24h', orientation: 'horizontal', theme: 'system', alwaysOnTop: false,
    layout: { digital: { width: 176, height: 92 }, analogue: { width: 176, height: 176 }, padding: 6 },
    refreshInMs: 60000, notices: [], scrolls: false, dragThreshold: { width: 4, height: 4 },
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
    MoveClock: vi.fn(async () => undefined),
    SearchPlaces: vi.fn(async () => places),
    SetStyle: vi.fn(async () => undefined),
    SetFormat: vi.fn(async () => undefined),
    SetOrientation: vi.fn(async () => undefined),
    SetTheme: vi.fn(async () => undefined),
    SetAlwaysOnTop: vi.fn(async () => undefined),
    StartWithWindows: vi.fn(async () => false),
    SetStartWithWindows: vi.fn(async () => undefined),
    DismissNotices: vi.fn(async () => undefined),
    ShowContextMenu: vi.fn(async () => undefined),
    OpenSettings: vi.fn(async () => undefined),
    CloseSettings: vi.fn(async () => undefined),
    Hide: vi.fn(async () => undefined),
    OpenDonation: vi.fn(async () => undefined),
  }
  window.go = { main: { App: bridge } }
  window.WailsInvoke = vi.fn()
  return bridge
}
