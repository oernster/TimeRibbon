// A stand-in for the Go facade in tests: every call is recorded, every answer is canned. The window's
// half is ribbonkit's own stand-in; TimeRibbon's methods are added to it here.

import { install, windowBridge } from '@oernster/ribbonkit/testing'
import { vi } from 'vitest'
import type { Cell, MenuChoice, Place, Snapshot } from './wire'

export { about } from '@oernster/ribbonkit/testing'

export function cell(overrides: Partial<Cell> = {}): Cell {
  return {
    id: 'ny', label: 'New York', zone: 'America/New_York', zoneMark: 'EDT', time: '16:37',
    date: 'Sunday, 27 September', hourAngle: 138.5, minuteAngle: 222, problem: '', ...overrides,
  }
}

function item(action: string, label: string, checked?: boolean): MenuChoice {
  return { action, label, checkable: checked != null, checked: checked === true, children: [] }
}

function group(label: string, children: MenuChoice[]): MenuChoice {
  return { action: '', label, checkable: false, checked: false, children }
}

/** The menus' choices as Go sends them, in their order; Colour cut to two schemes. */
export const choices: MenuChoice[] = [
  group('Style', [item('digital', 'Digital', true), item('analogue', 'Analogue', false)]),
  group('Colour', [item('colour-classic', 'Classic', true), item('colour-neon', 'Neon', false)]),
  group('Orientation', [item('horizontal', 'Horizontal', true), item('vertical', 'Vertical', false)]),
  group('Position', [item('top-edge', 'Centre on top edge'), item('bottom-edge', 'Centre on bottom edge')]),
  item('always-on-top', 'Always on top', false),
  item('pin', 'Pin ribbon', true),
  item('sun-map', 'Sun map', false),
]

export function snapshot(overrides: Partial<Snapshot> = {}): Snapshot {
  return {
    cells: [cell(), cell({ id: 'syd', label: 'Sydney', zone: 'Australia/Sydney', zoneMark: 'AEST', time: '06:37', date: 'Monday, 28 September' })],
    style: 'digital', size: 'large', colour: 'classic', format: '24h', dateFormat: 'day-month', orientation: 'horizontal', theme: 'system', alwaysOnTop: false, opacity: 100, minOpacity: 20, scale: 100, minScale: 75, maxScale: 200,
    layout: { digital: { width: 176, height: 92 }, analogue: { width: 176, height: 176 }, prompt: { width: 176, height: 184 }, padding: 6, handleLane: 16 },
    refreshInMs: 60000, notices: [], scrolls: false, dragThreshold: { width: 4, height: 4 }, startLabel: 'Start at sign-in',
    collapsed: false,
    sunMap: {
      on: false, pullOut: false, side: '', shown: false, ribbon: { x: 0, y: 0, width: 0, height: 0 },
      map: { x: 0, y: 0, width: 0, height: 0 }, latitude: 0, longitude: 0, marks: [],
    },
    choices,
    ...overrides,
  }
}

export const places: Place[] = [
  { zone: 'Asia/Kolkata', label: 'Kolkata', country: 'India' },
  { zone: 'Europe/Oslo', label: 'Oslo', country: 'Norway' },
]

/** installBridge puts a recording facade on window and answers it. */
export function installBridge() {
  return install({
    ...windowBridge(),
    Snapshot: vi.fn(async () => snapshot()),
    AddClock: vi.fn(async () => 'new'),
    RenameClock: vi.fn(async () => undefined),
    RezoneClock: vi.fn(async () => undefined),
    RemoveClock: vi.fn(async () => undefined),
    SearchPlaces: vi.fn(async () => places),
    SetSize: vi.fn(async () => undefined),
    SetFormat: vi.fn(async () => undefined),
    SetDateFormat: vi.fn(async () => undefined),
    DismissNotices: vi.fn(async () => undefined),
    TextSamples: vi.fn(async () => ({ times: ['00:00', '23:59'], dates: ['Friday, 1 May', 'Wednesday, 30 September'] })),
    SetMeasured: vi.fn(async () => undefined),
  })
}
