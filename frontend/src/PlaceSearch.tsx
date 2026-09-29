import { useEffect, useId, useRef, useState, type KeyboardEvent } from 'react'
import { api, type Place, type Refused } from './api'
import { ArtButton } from './ArtButton'

interface Props {
  /** heading names what choosing a place does, such as "Add a clock". */
  heading: string
  onChoose: (place: Place) => void
  refused: Refused
  /** autoFocus puts the cursor in the search as it appears. */
  autoFocus: boolean
  /**
   * onCancel gives the search a Cancel button, Escape calling it, for a search opened for one
   * purpose. Without it the search stays open (FR-626): Escape clears what was typed; with nothing
   * typed it goes on to whatever holds the search.
   */
  onCancel?: () => void
  /**
   * picture, for a search that stays open, is drawn beside the box (FR-626); pressing it chooses the
   * highlighted place; with nothing typed yet it puts the cursor in the box. Such a search lists
   * places only once something is typed.
   */
  picture?: { art: string; label: string }
}

/** The most results listed at once; typing narrows the rest. */
const shown = 50

/**
 * PlaceSearch finds a place by city, zone or country (FR-302). Up and Down move through the
 * results, Enter chooses, Escape cancels (NFR-U-3).
 */
export function PlaceSearch({ heading, onChoose, refused, autoFocus, onCancel, picture }: Props) {
  const [query, setQuery] = useState('')
  const [places, setPlaces] = useState<Place[]>([])
  const [active, setActive] = useState(0)
  const listId = useId()
  const box = useRef<HTMLInputElement>(null)
  const listed = picture == null || query !== ''

  const choose = (place: Place) => {
    onChoose(place)
    if (onCancel == null) {
      setQuery('')
    }
  }

  const pressed = () => {
    if (listed && places[active] != null) {
      choose(places[active])
    } else {
      box.current?.focus()
    }
  }

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
    } else if (event.key === 'Enter' && listed && places[active] != null) {
      event.preventDefault()
      choose(places[active])
    } else if (event.key === 'Escape' && (onCancel != null || query !== '')) {
      event.preventDefault()
      event.stopPropagation()
      if (onCancel != null) {
        onCancel()
      } else {
        setQuery('')
      }
    }
  }

  return (
    <section className="search" aria-label={heading}>
      <h2>{heading}</h2>
      <div className="search-row">
        <input
          ref={box}
          autoFocus={autoFocus}
          type="search"
          placeholder="City, zone or country"
          aria-label="Search places"
          aria-controls={listId}
          aria-activedescendant={listed && places[active] != null ? `${listId}-${active}` : undefined}
          value={query}
          onChange={(event) => setQuery(event.target.value)}
          onKeyDown={key}
        />
        {picture != null && <ArtButton art={picture.art} label={picture.label} large onClick={pressed} />}
      </div>
      <ul id={listId} role="listbox" aria-label="Places" hidden={!listed}>
        {listed && places.map((place, index) => (
          <li
            key={place.zone}
            id={`${listId}-${index}`}
            role="option"
            aria-selected={index === active}
            className={index === active ? 'active' : ''}
            onMouseEnter={() => setActive(index)}
            onClick={() => choose(place)}
          >
            <span className="label">{place.label}</span> <span className="muted">{place.country}</span>
            <span className="muted zone">{place.zone}</span>
          </li>
        ))}
        {listed && places.length === 0 && <li className="muted">No place matches</li>}
      </ul>
      {onCancel != null && (
        <button type="button" onClick={onCancel}>
          Cancel
        </button>
      )}
    </section>
  )
}
