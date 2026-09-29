// The wire between Go and the page, stated a second time here. dto.go is the other statement; a
// structural test compares the two.

export interface Size {
  width: number
  height: number
}

export interface Layout {
  digital: Size
  analogue: Size
  prompt: Size
  padding: number
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
  layout: Layout
  refreshInMs: number
  notices: string[]
  scrolls: boolean
  dragThreshold: Size
  startLabel: string
  /** True while the window is an unpinned ribbon's tab (FR-614). */
  collapsed: boolean
  sunMap: SunMap
}

/** A rectangle inside the window, in the window's pixels. */
export interface Box {
  x: number
  y: number
  width: number
  height: number
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

export interface About {
  name: string
  version: string
  author: string
  copyright: string
  credits: Credit[]
}

export interface UpdateStatus {
  current: string
  latest: string
  updateAvailable: boolean
}

export interface Credit {
  name: string
  licence: string
  role: string
}
