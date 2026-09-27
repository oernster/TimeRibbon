import { useEffect, useId, useState, type KeyboardEvent } from 'react'
import { api, type Place, type Refused } from './api'

interface Props {
  /** heading names what choosing a place does, such as "Add a clock". */
  heading: string
  onChoose: (place: Place) => void
  onCancel: () => void
  refused: Refused
}

/** The most results listed at once; typing narrows the rest. */
const shown = 50

/**
 * PlaceSearch finds a place by city, zone or country (FR-302). Up and Down move through the
 * results, Enter chooses, Escape cancels (NFR-U-3).
 */
export function PlaceSearch({ heading, onChoose, onCancel, refused }: Props) {
  const [query, setQuery] = useState('')
  const [places, setPlaces] = useState<Place[]>([])
  const [active, setActive] = useState(0)
  const listId = useId()

  useEffect(() => {
    let current = true
    void api.searchPlaces(query, refused).then((found) => {
      if (current && found != null) {
        setPlaces(found.slice(0, shown))
        setActive(0)
      }
    })
    return () => {
      current = false
    }
  }, [query, refused])

  const key = (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.key === 'ArrowDown') {
      event.preventDefault()
      setActive((index) => Math.min(index + 1, places.length - 1))
    } else if (event.key === 'ArrowUp') {
      event.preventDefault()
      setActive((index) => Math.max(index - 1, 0))
    } else if (event.key === 'Enter' && places[active] != null) {
      event.preventDefault()
      onChoose(places[active])
    } else if (event.key === 'Escape') {
      event.preventDefault()
      event.stopPropagation()
      onCancel()
    }
  }

  return (
    <section className="search" aria-label={heading}>
      <h2>{heading}</h2>
      <input
        autoFocus
        type="search"
        placeholder="City, zone or country"
        aria-label="Search places"
        aria-controls={listId}
        aria-activedescendant={places[active] != null ? `${listId}-${active}` : undefined}
        value={query}
        onChange={(event) => setQuery(event.target.value)}
        onKeyDown={key}
      />
      <ul id={listId} role="listbox" aria-label="Places">
        {places.map((place, index) => (
          <li
            key={place.zone}
            id={`${listId}-${index}`}
            role="option"
            aria-selected={index === active}
            className={index === active ? 'active' : ''}
            onMouseEnter={() => setActive(index)}
            onClick={() => onChoose(place)}
          >
            <span className="label">{place.label}</span> <span className="muted">{place.country}</span>
            <span className="muted zone">{place.zone}</span>
          </li>
        ))}
        {places.length === 0 && <li className="muted">No place matches</li>}
      </ul>
      <button type="button" onClick={onCancel}>
        Cancel
      </button>
    </section>
  )
}
