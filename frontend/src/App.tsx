import { useEffect } from 'react'
import { api } from './api'
import { useShell, type Panel } from '@oernster/ribbonkit'
import { About, Licence, Update } from './Help'
import { useMeasuredCells } from './measure'
import { Settings } from './Settings'
import { Surface } from './Surface'

/**
 * TimeRibbon's own open-panel word, which app.go sends: add-clock opens Settings on the place search.
 * The window's words are ribbonkit's (useShell).
 */
const addClock = 'add-clock'
const opens: Record<string, Panel> = { [addClock]: 'settings' }

/**
 * App is ribbonkit's shell around the clocks: it shows the ribbon, else the panel Go opened it at.
 * The snapshot is also taken again at each minute boundary Go names (FR-208).
 */
export function App() {
  const { snapshot, problem, refused, view, at, update, opened, load, openPanel, closePanel } = useShell({
    calls: api,
    take: api.snapshot,
    opens,
  })

  // Go widens the cells to the widest time and date the page really draws, which only it can measure.
  useMeasuredCells(snapshot, load, refused)

  useEffect(() => {
    if (snapshot == null) {
      return
    }
    const timer = window.setTimeout(load, snapshot.refreshInMs)
    return () => window.clearTimeout(timer)
  }, [snapshot, load])

  if (snapshot == null) {
    return <div className="problem">{problem}</div>
  }
  if (view === 'settings') {
    return <Settings snapshot={snapshot} startAdding={at === addClock} reload={load} onClose={closePanel} ready={opened} />
  }
  if (view === 'about') {
    return <About onClose={closePanel} />
  }
  if (view === 'licence') {
    return <Licence onClose={closePanel} />
  }
  if (view === 'update' && update != null) {
    return <Update status={update} onClose={closePanel} />
  }
  return <Surface snapshot={snapshot} onAddClock={() => openPanel(addClock)} refused={refused} />
}
