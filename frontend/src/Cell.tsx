import type { Cell as CellData } from './api'
import { Dial } from './Dial'

interface Props {
  cell: CellData
  analogue: boolean
}

/** Place names the clock: its label, then its zone mark (FR-203). The whole label is the tooltip. */
function Place({ cell }: { cell: CellData }) {
  return (
    <div className="place" title={cell.label}>
      <span className="label">{cell.label}</span>
      {cell.zoneMark !== '' && <span> · {cell.zoneMark}</span>}
    </div>
  )
}

/**
 * Cell is one clock. A clock that cannot be shown says why in words and shows no time, never
 * another zone's (FR-706, NFR-U-2).
 */
export function Cell({ cell, analogue }: Props) {
  if (cell.problem !== '') {
    return (
      <div className="cell" role="group" aria-label={cell.label}>
        <Place cell={cell} />
        <div className="problem">{cell.problem}</div>
      </div>
    )
  }
  if (analogue) {
    return (
      <div className="cell analogue" role="group" aria-label={`${cell.label}, ${cell.time}, ${cell.date}`}>
        <Place cell={cell} />
        <Dial hourAngle={cell.hourAngle} minuteAngle={cell.minuteAngle} />
        <div className="date">{cell.date}</div>
      </div>
    )
  }
  return (
    <div className="cell" role="group" aria-label={`${cell.label}, ${cell.time}, ${cell.date}`}>
      <Place cell={cell} />
      <div className="time">{cell.time}</div>
      <div className="date">{cell.date}</div>
    </div>
  )
}
