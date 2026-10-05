// TimeRibbon's half of the wire between Go and the page, stated a second time here. dto.go is the
// other statement; ribbonkit states the window's half. A structural test compares each pair.

import type { Box } from '@oernster/ribbonkit'

export interface Size {
  width: number
  height: number
}

export interface Layout {
  digital: Size
  analogue: Size
  prompt: Size
  padding: number
  handleLane: number
}

/** Every time and date a cell can show under the current formats, for the page to measure (FR-620). */
export interface TextSamples {
  times: string[]
  dates: string[]
}

/** The cell width the page measured its widest text to need, with the choices it measured under (FR-620). */
export interface Measured {
  size: string
  style: string
  format: string
  dateFormat: string
  cellWidth: number
}

export interface Cell {
  id: string
  label: string
  zone: string
  zoneMark: string
  time: string
  date: string
  hourAngle: number
  minuteAngle: number
  problem: string
}

export interface Snapshot {
  cells: Cell[]
  style: string
  size: string
  colour: string
  format: string
  dateFormat: string
  orientation: string
  theme: string
  alwaysOnTop: boolean
  /** How opaque the window is drawn in percent; minOpacity the least it may be (FR-622). */
  opacity: number
  minOpacity: number
  /** The percent the ribbon is drawn at on top of its size; minScale and maxScale bound it (FR-623). */
  scale: number
  minScale: number
  maxScale: number
  layout: Layout
  refreshInMs: number
  notices: string[]
  scrolls: boolean
  dragThreshold: Size
  startLabel: string
  /** True while the window is an unpinned ribbon's tab (FR-614). */
  collapsed: boolean
  sunMap: SunMap
  /** The menus' choices, which Settings offers as well (FR-624). */
  choices: MenuChoice[]
}

/** One of the menus' choices: either a group of children or one item whose action goes back to Choose. */
export interface MenuChoice {
  action: string
  label: string
  checkable: boolean
  checked: boolean
  /** Greyed, as a Position item that would leave the ribbon where it stands is (FR-408). */
  disabled: boolean
  children: MenuChoice[]
}

/** One clock's place on the sun map (FR-908). */
export interface Mark {
  label: string
  latitude: number
  longitude: number
}

/**
 * What the sun map draws and where (FR-901 to FR-910): side is where the map and the pull out's
 * handle go, empty while the map is off; shown is whether the map is drawn now, in map, beside the
 * ribbon in ribbon. latitude and longitude are the subsolar point.
 */
export interface SunMap {
  on: boolean
  pullOut: boolean
  side: string
  shown: boolean
  ribbon: Box
  map: Box
  latitude: number
  longitude: number
  marks: Mark[]
}

export interface Place {
  zone: string
  label: string
  country: string
}
