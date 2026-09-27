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
  format: string
  orientation: string
  theme: string
  alwaysOnTop: boolean
  layout: Layout
  refreshInMs: number
  notices: string[]
  scrolls: boolean
  dragThreshold: Size
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

export interface Credit {
  name: string
  licence: string
  role: string
}
