import { useCallback, useEffect, useState, type KeyboardEvent } from 'react'
import { api, type Cell, type Place, type Snapshot } from './api'
import { ClockList } from './ClockList'
import { PlaceSearch } from './PlaceSearch'

interface Props {
  snapshot: Snapshot
  /** startAdding opens straight onto the place search, as Add clock does. */
  startAdding: boolean
  reload: () => void
  onClose: () => void
}

type Search = { mode: 'add' } | { mode: 'rezone'; cell: Cell } | null

interface Choice {
  label: string
  value: string
}

/** Each choice Settings offers, with its values in the words shown (FR-601). */
const choices: { name: string; key: 'style' | 'format' | 'orientation' | 'theme'; options: Choice[] }[] = [
  { name: 'Style', key: 'style', options: [{ label: 'Digital', value: 'digital' }, { label: 'Analogue', value: 'analogue' }] },
  { name: 'Time format', key: 'format', options: [{ label: '24-hour', value: '24h' }, { label: '12-hour', value: '12h' }] },
  { name: 'Orientation', key: 'orientation', options: [{ label: 'Horizontal', value: 'horizontal' }, { label: 'Vertical', value: 'vertical' }] },
  { name: 'Theme', key: 'theme', options: [{ label: 'System', value: 'system' }, { label: 'Light', value: 'light' }, { label: 'Dark', value: 'dark' }] },
]

const setters = {
  style: api.setStyle,
  format: api.setFormat,
  orientation: api.setOrientation,
  theme: api.setTheme,
}

/**
 * Settings is the small set of choices plus the clocks (FR-601). Every change applies and is kept
 * at once, with no Save step (FR-602). Escape closes it.
 */
export function Settings({ snapshot, startAdding, reload, onClose }: Props) {
  const [search, setSearch] = useState<Search>(startAdding ? { mode: 'add' } : null)
  const [problem, setProblem] = useState('')
  const [startWithWindows, setStartWithWindows] = useState<boolean | null>(null)

  useEffect(() => {
    void api.startWithWindows(setProblem).then(setStartWithWindows)
  }, [])

  const then = useCallback(() => reload(), [reload])

  const chosen = (place: Place) => {
    if (search?.mode === 'rezone') {
      void api.rezoneClock(search.cell.id, place.zone, setProblem).then(then)
    } else {
      void api.addClock(place.zone, setProblem).then(then)
    }
    setSearch(null)
  }

  const escape = (event: KeyboardEvent<HTMLElement>) => {
    if (event.key === 'Escape' && search == null) {
      onClose()
    }
  }

  return (
    <main className="settings" onKeyDown={escape}>
      <header>
        <h1>Settings</h1>
        <button type="button" onClick={onClose}>
          Close
        </button>
      </header>
      {problem !== '' && (
        <p className="problem" role="alert">
          {problem}
        </p>
      )}

      <h2>Clocks</h2>
      <ClockList
        cells={snapshot.cells}
        onRename={(id, label) => void api.renameClock(id, label, setProblem).then(then)}
        onChangePlace={(cell) => setSearch({ mode: 'rezone', cell })}
        onMove={(id, steps) => void api.moveClock(id, steps, setProblem).then(then)}
        onRemove={(id) => void api.removeClock(id, setProblem).then(then)}
      />
      {search == null ? (
        <button type="button" onClick={() => setSearch({ mode: 'add' })}>
          Add clock
        </button>
      ) : (
        <PlaceSearch
          heading={search.mode === 'rezone' ? `Change the place of ${search.cell.label}` : 'Add a clock'}
          onChoose={chosen}
          onCancel={() => setSearch(null)}
          refused={setProblem}
        />
      )}

      {choices.map((choice) => (
        <fieldset key={choice.key}>
          <legend>{choice.name}</legend>
          {choice.options.map((option) => (
            <label key={option.value}>
              <input
                type="radio"
                name={choice.key}
                value={option.value}
                checked={snapshot[choice.key] === option.value}
                onChange={() => void setters[choice.key](option.value, setProblem).then(then)}
              />
              {option.label}
            </label>
          ))}
        </fieldset>
      ))}

      <fieldset>
        <legend>Window</legend>
        <label>
          <input
            type="checkbox"
            checked={snapshot.alwaysOnTop}
            onChange={(event) => void api.setAlwaysOnTop(event.target.checked, setProblem).then(then)}
          />
          Always on top
        </label>
        <label>
          <input
            type="checkbox"
            disabled={startWithWindows == null}
            checked={startWithWindows === true}
            onChange={(event) => {
              const on = event.target.checked
              void api.setStartWithWindows(on, setProblem).then(() => api.startWithWindows(setProblem).then(setStartWithWindows))
            }}
          />
          Start with Windows
        </label>
      </fieldset>
    </main>
  )
}
