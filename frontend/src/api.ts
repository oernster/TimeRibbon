// Typed access to the Go facade. Wails injects window.go.main.App and window.runtime at load time;
// wrapping them here keeps the binding shape in one file.
//
// Every call that Go can refuse takes a refusal handler as its last argument and answers null
// rather than rejecting, so a call without a handler does not compile (ported from Bridge Talk).

import type { About, Measured, Place, Snapshot, TextSamples } from './wire'

export type { About, Box, Cell, Credit, Layout, Mark, Measured, MenuChoice, Place, Size, Snapshot, SunMap, TextSamples, UpdateStatus } from './wire'

/** A handler told, in words, why a call was refused. */
export type Refused = (reason: string) => void

interface Bridge {
  Snapshot(): Promise<Snapshot>
  AddClock(zone: string): Promise<string>
  RenameClock(id: string, label: string): Promise<void>
  RezoneClock(id: string, zone: string): Promise<void>
  RemoveClock(id: string): Promise<void>
  SearchPlaces(query: string): Promise<Place[]>
  SetSize(size: string): Promise<void>
  SetFormat(format: string): Promise<void>
  SetDateFormat(dateFormat: string): Promise<void>
  SetTheme(theme: string): Promise<void>
  StartWithWindows(): Promise<boolean>
  SetStartWithWindows(on: boolean): Promise<void>
  DismissNotices(): Promise<void>
  SetScrollbar(dip: number): Promise<void>
  SetOpacity(percent: number): Promise<void>
  BeginScale(thickness: number, x: number, y: number): Promise<void>
  DragScale(x: number, y: number): Promise<void>
  EndScale(x: number, y: number): Promise<void>
  SetScale(percent: number): Promise<void>
  TextSamples(): Promise<TextSamples>
  SetMeasured(measured: Measured): Promise<void>
  SetPixelRatio(ratio: number): Promise<void>
  SetBackground(red: number, green: number, blue: number): Promise<void>
  RibbonDrawn(): Promise<void>
  TogglePullOut(): Promise<void>
  ShowContextMenu(): Promise<void>
  Choose(action: string): Promise<void>
  OpenPanel(panel: string): Promise<void>
  FitPanel(height: number): Promise<void>
  ClosePanel(): Promise<void>
  Hide(): Promise<void>
  OpenDonation(): Promise<void>
  OpenUpdate(): Promise<void>
  SkipUpdate(): Promise<void>
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

const unreachable = 'TimeRibbon is not running behind this page'

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
  setSize: (size: string, refused: Refused) => call((b) => b.SetSize(size), refused),
  setFormat: (format: string, refused: Refused) => call((b) => b.SetFormat(format), refused),
  setDateFormat: (dateFormat: string, refused: Refused) => call((b) => b.SetDateFormat(dateFormat), refused),
  setTheme: (theme: string, refused: Refused) => call((b) => b.SetTheme(theme), refused),
  startWithWindows: (refused: Refused) => call((b) => b.StartWithWindows(), refused),
  setStartWithWindows: (on: boolean, refused: Refused) => call((b) => b.SetStartWithWindows(on), refused),
  dismissNotices: (refused: Refused) => call((b) => b.DismissNotices(), refused),
  setScrollbar: (dip: number, refused: Refused) => call((b) => b.SetScrollbar(dip), refused),
  setOpacity: (percent: number, refused: Refused) => call((b) => b.SetOpacity(percent), refused),
  beginScale: (thickness: number, x: number, y: number, refused: Refused) =>
    call((b) => b.BeginScale(thickness, x, y), refused),
  dragScale: (x: number, y: number, refused: Refused) => call((b) => b.DragScale(x, y), refused),
  endScale: (x: number, y: number, refused: Refused) => call((b) => b.EndScale(x, y), refused),
  setScale: (percent: number, refused: Refused) => call((b) => b.SetScale(percent), refused),
  textSamples: (refused: Refused) => call((b) => b.TextSamples(), refused),
  setMeasured: (measured: Measured, refused: Refused) => call((b) => b.SetMeasured(measured), refused),
  setPixelRatio: (ratio: number, refused: Refused) => call((b) => b.SetPixelRatio(ratio), refused),
  setBackground: (red: number, green: number, blue: number, refused: Refused) =>
    call((b) => b.SetBackground(red, green, blue), refused),
  ribbonDrawn: (refused: Refused) => call((b) => b.RibbonDrawn(), refused),
  togglePullOut: (refused: Refused) => call((b) => b.TogglePullOut(), refused),
  showContextMenu: (refused: Refused) => call((b) => b.ShowContextMenu(), refused),
  choose: (action: string, refused: Refused) => call((b) => b.Choose(action), refused),
  openPanel: (panel: string, refused: Refused) => call((b) => b.OpenPanel(panel), refused),
  fitPanel: (height: number, refused: Refused) => call((b) => b.FitPanel(height), refused),
  closePanel: (refused: Refused) => call((b) => b.ClosePanel(), refused),
  hide: (refused: Refused) => call((b) => b.Hide(), refused),
  openDonation: (refused: Refused) => call((b) => b.OpenDonation(), refused),
  openUpdate: (refused: Refused) => call((b) => b.OpenUpdate(), refused),
  skipUpdate: (refused: Refused) => call((b) => b.SkipUpdate(), refused),
  about: (refused: Refused) => call((b) => b.About(), refused),
  licence: (refused: Refused) => call((b) => b.Licence(), refused),
}

/** on listens for a Go event, answering the call that stops listening. Outside Wails it hears nothing. */
export function on(name: string, callback: (...data: unknown[]) => void): () => void {
  return window.runtime?.EventsOn(name, callback) ?? (() => undefined)
}

/** startDrag hands the press to the system's own window drag, as Wails' drag regions do (FR-401). */
export function startDrag(): void {
  window.WailsInvoke?.('drag')
}
