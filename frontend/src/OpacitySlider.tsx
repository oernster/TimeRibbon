import { useState } from 'react'
import { api, type Refused, type Snapshot } from './api'
import { percentOfWhole } from '@oernster/ribbonkit'

/** The slider moves in whole steps of this many percent. */
const step = 5

interface Props {
  snapshot: Snapshot
  refused: Refused
  then: () => void
}

/**
 * OpacitySlider chooses how opaque the ribbon is drawn, from the least the setting allows to wholly
 * opaque (FR-622). The panel it sits in stays opaque, so the slider shows the value as it moves; the
 * choice is kept once it is let go, so a drag is one change rather than one for every step it passes.
 */
export function OpacitySlider({ snapshot, refused, then }: Props) {
  const [moving, setMoving] = useState<number | null>(null)
  const shown = moving ?? snapshot.opacity
  const keep = () => {
    if (moving == null) {
      return
    }
    void api.setOpacity(moving, refused).then(() => {
      setMoving(null)
      then()
    })
  }
  return (
    <fieldset>
      <legend>Opacity</legend>
      <label>
        <input
          type="range"
          min={snapshot.minOpacity}
          max={percentOfWhole}
          step={step}
          value={shown}
          aria-valuetext={`${shown} percent`}
          onChange={(event) => setMoving(Number(event.target.value))}
          onPointerUp={keep}
          onKeyUp={keep}
          onBlur={keep}
        />
        <span className="opacity-value">{shown}%</span>
      </label>
    </fieldset>
  )
}
