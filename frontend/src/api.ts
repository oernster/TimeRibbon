// Typed access to the Go facade: ribbonkit's window calls plus TimeRibbon's own, over one guarded
// call (ribbonkit's bridge). Wrapping them here keeps the binding shape in one file.

import { connect, windowCalls, type Refused, type WindowBridge } from '@oernster/ribbonkit'
import type { Measured, Place, Snapshot, TextSamples } from './wire'

export { on, startDrag, type Credit, type Refused, type UpdateStatus } from '@oernster/ribbonkit'
export type { Cell, Layout, Mark, Measured, Place, Size, Snapshot, SunMap, TextSamples } from './wire'

/** Bridge is the App Wails binds: the window's methods (WindowBridge) and TimeRibbon's own (app.go). */
interface Bridge extends WindowBridge {
  Snapshot(): Promise<Snapshot>
  AddClock(zone: string): Promise<string>
  RenameClock(id: string, label: string): Promise<void>
  RezoneClock(id: string, zone: string): Promise<void>
  RemoveClock(id: string): Promise<void>
  SearchPlaces(query: string): Promise<Place[]>
  SetSize(size: string): Promise<void>
  SetFormat(format: string): Promise<void>
  SetDateFormat(dateFormat: string): Promise<void>
  DismissNotices(): Promise<void>
  TextSamples(): Promise<TextSamples>
  SetMeasured(measured: Measured): Promise<void>
}

const call = connect<Bridge>('TimeRibbon')

export const api = {
  ...windowCalls(call),
  snapshot: (refused: Refused) => call((b) => b.Snapshot(), refused),
  addClock: (zone: string, refused: Refused) => call((b) => b.AddClock(zone), refused),
  renameClock: (id: string, label: string, refused: Refused) => call((b) => b.RenameClock(id, label), refused),
  rezoneClock: (id: string, zone: string, refused: Refused) => call((b) => b.RezoneClock(id, zone), refused),
  removeClock: (id: string, refused: Refused) => call((b) => b.RemoveClock(id), refused),
  searchPlaces: (query: string, refused: Refused) => call((b) => b.SearchPlaces(query), refused),
  setSize: (size: string, refused: Refused) => call((b) => b.SetSize(size), refused),
  setFormat: (format: string, refused: Refused) => call((b) => b.SetFormat(format), refused),
  setDateFormat: (dateFormat: string, refused: Refused) => call((b) => b.SetDateFormat(dateFormat), refused),
  dismissNotices: (refused: Refused) => call((b) => b.DismissNotices(), refused),
  textSamples: (refused: Refused) => call((b) => b.TextSamples(), refused),
  setMeasured: (measured: Measured, refused: Refused) => call((b) => b.SetMeasured(measured), refused),
}
