import { useEffect } from 'react'
import { api, type Refused, type Snapshot, type TextSamples } from './api'

/**
 * widestLine answers the width of the widest of texts laid out with className inside parent: one
 * box that shrinks to its widest line, so a single layout pass measures them all.
 */
function widestLine(parent: HTMLElement, className: string, texts: string[]): number {
  const box = parent.ownerDocument.createElement('div')
  box.style.cssText = 'position: absolute; width: max-content'
  for (const text of texts) {
    const line = parent.ownerDocument.createElement('div')
    line.className = className
    line.textContent = text
    box.appendChild(line)
  }
  parent.appendChild(box)
  return box.getBoundingClientRect().width
}

/** sides answers the sum of a computed style's two horizontal lengths named by prefix and suffix. */
function sides(style: CSSStyleDeclaration, prefix: string, suffix: string): number {
  return (parseFloat(style.getPropertyValue(`${prefix}-left${suffix}`)) || 0) + (parseFloat(style.getPropertyValue(`${prefix}-right${suffix}`)) || 0)
}

/**
 * cellWidthNeeded measures, in DIP, how wide a cell must be to show every sample whole in the font
 * this web engine really draws with (FR-620): the widest time (a digital cell's only) and the widest
 * date, each in the classes a cell's own text carries at size, plus the cell's padding and border.
 * It measures a cell that follows another, so the divider between cells is counted.
 */
export function cellWidthNeeded(samples: TextSamples, size: string, analogue: boolean, doc: Document = document): number {
  const ribbon = doc.createElement('div')
  ribbon.className = `ribbon horizontal ${size}`
  ribbon.style.cssText = 'position: absolute; visibility: hidden; left: 0; top: 0; width: auto; height: auto'
  const first = doc.createElement('div')
  const cell = doc.createElement('div')
  first.className = cell.className = analogue ? 'cell analogue' : 'cell'
  ribbon.append(first, cell)
  doc.body.appendChild(ribbon)
  try {
    const text = Math.max(analogue ? 0 : widestLine(cell, 'time', samples.times), widestLine(cell, 'date', samples.dates))
    const style = doc.defaultView?.getComputedStyle(cell)
    const chrome = style == null ? 0 : sides(style, 'padding', '') + sides(style, 'border', '-width')
    return Math.ceil(text + chrome)
  } finally {
    ribbon.remove()
  }
}

/**
 * useMeasuredCells tells Go how wide a cell must be for its widest time and date whenever a choice
 * that changes the text or its font does, then takes the snapshot again so the cells are drawn at
 * the width Go settles on (FR-620). A measurement overtaken by a newer choice is dropped.
 */
export function useMeasuredCells(snapshot: Snapshot | null, load: () => void, refused: Refused): void {
  const size = snapshot?.size
  const style = snapshot?.style
  const format = snapshot?.format
  const dateFormat = snapshot?.dateFormat
  useEffect(() => {
    if (size == null || style == null || format == null || dateFormat == null) {
      return
    }
    let overtaken = false
    const measure = async () => {
      await document.fonts?.ready
      const samples = await api.textSamples(refused)
      if (samples == null || overtaken) {
        return
      }
      const cellWidth = cellWidthNeeded(samples, size, style === 'analogue')
      const taken = await api.setMeasured({ size, style, format, dateFormat, cellWidth }, refused)
      if (taken === null || overtaken) {
        return
      }
      load()
    }
    void measure()
    return () => {
      overtaken = true
    }
  }, [size, style, format, dateFormat, load, refused])
}
