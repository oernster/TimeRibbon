// The wire between Go and the page, stated a second time here. dto.go is the other statement; a
// structural test compares the two.

export interface Size {
  width: number
  height: number
}

export interface Layout {
  digital: Size
  analogue: Size
  padding: number
}

export interface Cell {
  id: string
  label: string
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
