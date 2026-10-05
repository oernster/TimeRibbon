import { api, type Refused, type Snapshot } from './api'
import { PullOut } from '@oernster/ribbonkit'
import { Ribbon } from './Ribbon'
import { SunMap } from './SunMap'

interface Props {
  snapshot: Snapshot
  onAddClock: () => void
  refused: Refused
}

/** The handle's words, one home: the button says what pressing it will do (NFR-U-4). */
export const openPullOut = 'Open the sun map'
export const closePullOut = 'Close the sun map'

/**
 * Surface is the window's content while it is the ribbon: the ribbon alone, else the ribbon with its
 * sun map pulled out beside it by ribbonkit's handle (FR-902, FR-903).
 */
export function Surface({ snapshot, onAddClock, refused }: Props) {
  const map = snapshot.sunMap
  const ribbon = <Ribbon snapshot={snapshot} onAddClock={onAddClock} refused={refused} />
  if (snapshot.collapsed || map.side === '') {
    return ribbon
  }
  return (
    <PullOut
      side={map.side}
      ribbon={map.ribbon}
      beside={map.shown ? map.map : null}
      pulledOut={map.pullOut}
      scale={snapshot.scale}
      handleLane={snapshot.layout.handleLane}
      dragThreshold={snapshot.dragThreshold}
      calls={api}
      refused={refused}
      words={{ open: openPullOut, close: closePullOut }}
      band={ribbon}
    >
      <SunMap sunMap={map} width={map.map.width} height={map.map.height} dragThreshold={snapshot.dragThreshold} refused={refused} />
    </PullOut>
  )
}
