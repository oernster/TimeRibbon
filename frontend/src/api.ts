// Typed access to the Go facade. Wails injects window.go.main.App and window.runtime at load time;
// wrapping them here keeps the binding shape in one file.
//
// Every call that Go can refuse takes a refusal handler as its last argument and answers null
// rather than rejecting, so a call without a handler does not compile (ported from Bridge Talk).

import type { About, Place, Snapshot } from './wire'

export type { About, Cell, Credit, Layout, Place, Size, Snapshot } from './wire'

/** A handler told, in words, why a call was refused. */
export type Refused = (reason: string) => void

interface Bridge {
  Snapshot(): Promise<Snapshot>
  AddClock(zone: string): Promise<string>
  RenameClock(id: string, label: string): Promise<void>
  RezoneClock(id: string, zone: string): Promise<void>
  RemoveClock(id: string): Promise<void>
  SearchPlaces(query: string): Promise<Place[]>
  SetStyle(style: string): Promise<void>
  SetSize(size: string): Promise<void>
  SetFormat(format: string): Promise<void>
  SetOrientation(orientation: string): Promise<void>
  SetTheme(theme: string): Promise<void>
  SetAlwaysOnTop(on: boolean): Promise<void>
  StartWithWindows(): Promise<boolean>
  SetStartWithWindows(on: boolean): Promise<void>
  DismissNotices(): Promise<void>
  SetScrollbar(dip: number): Promise<void>
  ShowContextMenu(): Promise<void>
  OpenPanel(): Promise<void>
  ClosePanel(): Promise<void>
  Hide(): Promise<void>
  OpenDonation(): Promise<void>
  About(): Promise<About>
  Licence(): Promise<string>
}

interface Runtime {
  EventsOn(name: string, callback: (...data: unknown[]) => void): () => void
}

declare global {
  interface Window {
    go?: { main?: { App?: Bridge } }
    runtime?: Runtime
    WailsInvoke?: (message: string) => void
  }
}

const unreachable = 'TimeStrip is not running behind this page'

/** call runs one facade call, answering null and telling refused why when it cannot. */
async function call<T>(run: (bridge: Bridge) => Promise<T>, refused: Refused): Promise<T | null> {
  const bridge = window.go?.main?.App
  if (bridge == null) {
    refused(unreachable)
    return null
  }
  try {
    return await run(bridge)
  } catch (failure) {
    refused(String(failure))
    return null
  }
}

export const api = {
  snapshot: (refused: Refused) => call((b) => b.Snapshot(), refused),
  addClock: (zone: string, refused: Refused) => call((b) => b.AddClock(zone), refused),
  renameClock: (id: string, label: string, refused: Refused) => call((b) => b.RenameClock(id, label), refused),
  rezoneClock: (id: string, zone: string, refused: Refused) => call((b) => b.RezoneClock(id, zone), refused),
  removeClock: (id: string, refused: Refused) => call((b) => b.RemoveClock(id), refused),
  searchPlaces: (query: string, refused: Refused) => call((b) => b.SearchPlaces(query), refused),
  setStyle: (style: string, refused: Refused) => call((b) => b.SetStyle(style), refused),
  setSize: (size: string, refused: Refused) => call((b) => b.SetSize(size), refused),
  setFormat: (format: string, refused: Refused) => call((b) => b.SetFormat(format), refused),
  setOrientation: (orientation: string, refused: Refused) => call((b) => b.SetOrientation(orientation), refused),
  setTheme: (theme: string, refused: Refused) => call((b) => b.SetTheme(theme), refused),
  setAlwaysOnTop: (on: boolean, refused: Refused) => call((b) => b.SetAlwaysOnTop(on), refused),
  startWithWindows: (refused: Refused) => call((b) => b.StartWithWindows(), refused),
  setStartWithWindows: (on: boolean, refused: Refused) => call((b) => b.SetStartWithWindows(on), refused),
  dismissNotices: (refused: Refused) => call((b) => b.DismissNotices(), refused),
  setScrollbar: (dip: number, refused: Refused) => call((b) => b.SetScrollbar(dip), refused),
  showContextMenu: (refused: Refused) => call((b) => b.ShowContextMenu(), refused),
  openPanel: (refused: Refused) => call((b) => b.OpenPanel(), refused),
  closePanel: (refused: Refused) => call((b) => b.ClosePanel(), refused),
  hide: (refused: Refused) => call((b) => b.Hide(), refused),
  openDonation: (refused: Refused) => call((b) => b.OpenDonation(), refused),
  about: (refused: Refused) => call((b) => b.About(), refused),
  licence: (refused: Refused) => call((b) => b.Licence(), refused),
}

/** on listens for a Go event, answering the call that stops listening. Outside Wails it hears nothing. */
export function on(name: string, callback: (...data: unknown[]) => void): () => void {
  return window.runtime?.EventsOn(name, callback) ?? (() => undefined)
}

/** startDrag hands the press to Windows' own move loop, as Wails' drag regions do (FR-401). */
export function startDrag(): void {
  window.WailsInvoke?.('drag')
}
