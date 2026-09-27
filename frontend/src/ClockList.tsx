import { useState } from 'react'
import type { Cell } from './api'

interface Props {
  cells: Cell[]
  onRename: (id: string, label: string) => void
  onChangePlace: (cell: Cell) => void
  onMove: (id: string, steps: number) => void
  onRemove: (id: string) => void
}

/**
 * ClockList is the configured clocks in order (FR-303 to FR-306). Rows reorder by dragging or by
 * Move up and Move down; Remove asks first, naming the clock.
 */
export function ClockList({ cells, onRename, onChangePlace, onMove, onRemove }: Props) {
  const [dragged, setDragged] = useState<number | null>(null)
  const [confirming, setConfirming] = useState<Cell | null>(null)

  if (cells.length === 0) {
    return <p className="muted">No clocks yet. Add one below.</p>
  }
  return (
    <>
      <ol className="clocks" aria-label="Clocks">
        {cells.map((cell, index) => (
          <li
            key={cell.id}
            draggable
            onDragStart={() => setDragged(index)}
            onDragOver={(event) => event.preventDefault()}
            onDrop={() => {
              if (dragged != null && dragged !== index) {
                onMove(cells[dragged].id, index - dragged)
              }
              setDragged(null)
            }}
            onDragEnd={() => setDragged(null)}
          >
            <span className="grip" aria-hidden="true">
              ⋮⋮
            </span>
            <div className="row-main">
              <input
                key={cell.label}
                aria-label={`Label for ${cell.zone}`}
                defaultValue={cell.label}
                maxLength={32}
                onBlur={(event) => {
                  if (event.target.value !== cell.label) {
                    onRename(cell.id, event.target.value)
                  }
                }}
                onKeyDown={(event) => {
                  if (event.key === 'Enter') {
                    event.currentTarget.blur()
                  }
                }}
              />
              <span className={cell.problem !== '' ? 'problem' : 'muted zone'}>
                {cell.problem !== '' ? cell.problem : cell.zone}
              </span>
            </div>
            <button type="button" onClick={() => onChangePlace(cell)} title="Change place">
              Change place
            </button>
            <button type="button" aria-label={`Move ${cell.label} up`} title="Move up" disabled={index === 0} onClick={() => onMove(cell.id, -1)}>
              ↑
            </button>
            <button
              type="button"
              aria-label={`Move ${cell.label} down`}
              title="Move down"
              disabled={index === cells.length - 1}
              onClick={() => onMove(cell.id, 1)}
            >
              ↓
            </button>
            <button type="button" aria-label={`Remove ${cell.label}`} title="Remove" onClick={() => setConfirming(cell)}>
              Remove
            </button>
          </li>
        ))}
      </ol>
      {confirming != null && (
        <div className="dialog" role="alertdialog" aria-modal="true" aria-label="Remove clock">
          <p>Remove the clock for {confirming.label}?</p>
          <button
            type="button"
            autoFocus
            onClick={() => {
              onRemove(confirming.id)
              setConfirming(null)
            }}
          >
            Remove
          </button>
          <button type="button" onClick={() => setConfirming(null)}>
            Cancel
          </button>
        </div>
      )}
    </>
  )
}
