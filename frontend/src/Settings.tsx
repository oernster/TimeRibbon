import { useCallback, useEffect, useState, type KeyboardEvent } from 'react'
import { api, type Cell, type MenuChoice, type Place, type Refused, type Snapshot } from './api'
import { ClockList } from './ClockList'
import { PlaceSearch } from './PlaceSearch'
import { ArtButton, addClockTip } from './ArtButton'
import addClockArt from './assets/add-clock.png'
import donateMark from './assets/donate.png'
import { OpacitySlider, usePanelFit } from '@oernster/ribbonkit'

/** The picture alone does not say pressing it leaves the application, so the tip does. */
export const donateTip = 'Buy the author a drink (opens your browser)'

interface Props {
  snapshot: Snapshot
  /** startAdding puts the cursor in the place search, as Add clock does. */
  startAdding: boolean
  reload: () => void
  onClose: () => void
  /** ready is true once Go has made the window this panel, so its content can be measured. */
  ready?: boolean
}

interface Choice {
  label: string
  value: string
}

/**
 * The choices Settings alone offers, with their values in the words shown (FR-601). The menus'
 * choices come from Go in the snapshot, so their words have one home (FR-624).
 */
const choices: { name: string; key: 'size' | 'format' | 'dateFormat' | 'theme'; options: Choice[] }[] = [
  { name: 'Size', key: 'size', options: [{ label: 'Large', value: 'large' }, { label: 'Small', value: 'small' }] },
  { name: 'Time format', key: 'format', options: [{ label: '24-hour', value: '24h' }, { label: '12-hour', value: '12h' }] },
  {
    // FR-612: each labelled by its order, so no sample date goes stale.
    name: 'Date format',
    key: 'dateFormat',
    options: [
      { label: '28 September', value: 'day-month' },
      { label: 'September 28', value: 'month-day' },
      { label: 'DD/MM/YYYY', value: 'dmy' },
      { label: 'MM/DD/YYYY', value: 'mdy' },
      { label: 'YYYY/MM/DD', value: 'ymd' },
    ],
  },
  { name: 'Theme', key: 'theme', options: [{ label: 'System', value: 'system' }, { label: 'Light', value: 'light' }, { label: 'Dark', value: 'dark' }] },
]

const setters = {
  size: api.setSize,
  format: api.setFormat,
  dateFormat: api.setDateFormat,
  theme: api.setTheme,
}

interface ChoiceProps {
  choice: MenuChoice
  choose: (action: string) => void
}

/**
 * MenuGroup draws one of the menus' submenus (FR-624): one choice among ticked items as radio
 * buttons, a set of moves such as Position as plain buttons; a disabled item greyed (FR-408).
 */
function MenuGroup({ choice, choose }: ChoiceProps) {
  return (
    <fieldset>
      <legend>{choice.label}</legend>
      {choice.children.map((item) =>
        item.checkable ? (
          <label key={item.action}>
            <input type="radio" name={choice.label} value={item.action} checked={item.checked} disabled={item.disabled} onChange={() => choose(item.action)} />
            {item.label}
          </label>
        ) : (
          <button key={item.action} type="button" disabled={item.disabled} onClick={() => choose(item.action)}>
            {item.label}
          </button>
        ),
      )}
    </fieldset>
  )
}

/** MenuToggle draws one of the menus' ticked items that stands alone, such as Pin ribbon (FR-624). */
function MenuToggle({ choice, choose }: ChoiceProps) {
  return (
    <label>
      <input type="checkbox" checked={choice.checked} onChange={() => choose(choice.action)} />
      {choice.label}
    </label>
  )
}

/**
 * Settings is the clocks plus every choice, the menus' included (FR-601, FR-624). Every change
 * applies and is kept at once, with no Save step (FR-602). Escape closes it.
 */
export function Settings({ snapshot, startAdding, reload, onClose, ready = true }: Props) {
  const [rezoning, setRezoning] = useState<Cell | null>(null)
  const [problem, setProblem] = useState('')
  const [startWithWindows, setStartWithWindows] = useState<boolean | null>(null)
  const panel = usePanelFit<HTMLElement>(api, setProblem, ready)

  useEffect(() => {
    void api.startWithWindows(setProblem).then(setStartWithWindows)
  }, [])

  const then = useCallback(() => reload(), [reload])
  const refused: Refused = setProblem
  const choose = (action: string) => void api.choose(action, refused).then(then)

  const added = (place: Place) => void api.addClock(place.zone, refused).then(then)

  const rezoned = (place: Place) => {
    if (rezoning != null) {
      void api.rezoneClock(rezoning.id, place.zone, refused).then(then)
    }
    setRezoning(null)
  }

  const escape = (event: KeyboardEvent<HTMLElement>) => {
    if (event.key === 'Escape') {
      onClose()
    }
  }

  const groups = snapshot.choices.filter((choice) => choice.children.length > 0)
  const toggles = snapshot.choices.filter((choice) => choice.children.length === 0)

  return (
    <main className="settings" ref={panel} onKeyDown={escape}>
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
        onRename={(id, label) => void api.renameClock(id, label, refused).then(then)}
        onChangePlace={setRezoning}
        onRemove={(id) => void api.removeClock(id, refused).then(then)}
      />
      {rezoning == null ? (
        <PlaceSearch
          key="add"
          heading="Add a clock"
          onChoose={added}
          refused={refused}
          autoFocus={startAdding}
          picture={{ art: addClockArt, label: addClockTip }}
        />
      ) : (
        <PlaceSearch
          key={`rezone-${rezoning.id}`}
          heading={`Change the place of ${rezoning.label}`}
          onChoose={rezoned}
          onCancel={() => setRezoning(null)}
          refused={refused}
          autoFocus
        />
      )}

      <div className="settings-choices">
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
                  onChange={() => void setters[choice.key](option.value, refused).then(then)}
                />
                {option.label}
              </label>
            ))}
          </fieldset>
        ))}

        {groups.map((choice) => (
          <MenuGroup key={choice.label} choice={choice} choose={choose} />
        ))}

        <OpacitySlider opacity={snapshot.opacity} minOpacity={snapshot.minOpacity} calls={api} refused={refused} then={then} />

        <fieldset>
          <legend>Window</legend>
          {toggles.map((choice) => (
            <MenuToggle key={choice.action} choice={choice} choose={choose} />
          ))}
          <label>
            <input
              type="checkbox"
              disabled={startWithWindows == null}
              checked={startWithWindows === true}
              onChange={(event) => {
                const on = event.target.checked
                void api.setStartWithWindows(on, refused).then(() => api.startWithWindows(refused).then(setStartWithWindows))
              }}
            />
            {snapshot.startLabel}
          </label>
        </fieldset>
      </div>

      <footer className="settings-foot">
        <ArtButton art={donateMark} label={donateTip} onClick={() => void api.openDonation(refused)} />
        <p className="muted">Free to use and staying free: nothing is held back behind a donation.</p>
      </footer>
    </main>
  )
}
