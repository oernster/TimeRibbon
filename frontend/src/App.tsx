import { useCallback, useEffect, useState } from 'react'
import { api, on, type Snapshot } from './api'
import { About, Licence } from './Help'
import { Settings } from './Settings'
import { Strip } from './Strip'

/** The panels the window can become (CON-6); app.go names each in its open-panel event. */
type Panel = 'settings' | 'about' | 'licence'
type View = 'strip' | Panel

/** The open-panel event's words, each naming the panel it opens; add-clock opens Settings on the place search. */
const addClock = 'add-clock'
const panelFor: Record<string, Panel> = { settings: 'settings', [addClock]: 'settings', about: 'about', licence: 'licence' }

/**
 * App holds the snapshot and which surface the window shows. The snapshot is taken again at each
 * minute boundary Go names (FR-208) and whenever Go says the time, the displays or a choice changed
 * (FR-209).
 */
export function App() {
  const [snapshot, setSnapshot] = useState<Snapshot | null>(null)
  const [problem, setProblem] = useState('')
  const [view, setView] = useState<View>('strip')
  const [adding, setAdding] = useState(false)

  const load = useCallback(() => {
    void api.snapshot(setProblem).then((next) => {
      if (next != null) {
        setSnapshot(next)
        setProblem('')
      }
    })
  }, [])

  const openPanel = useCallback((at?: unknown) => {
    setAdding(at === addClock)
    setView(panelFor[String(at)] ?? 'settings')
    void api.openPanel(setProblem)
  }, [])

  const closePanel = useCallback(() => {
    setView('strip')
    void api.closePanel(setProblem).then(load)
  }, [load])

  useEffect(() => {
    load()
    const stopRefresh = on('refresh', load)
    const stopPanel = on('open-panel', openPanel)
    return () => {
      stopRefresh()
      stopPanel()
    }
  }, [load, openPanel])

  useEffect(() => {
    if (snapshot == null) {
      return
    }
    const timer = window.setTimeout(load, snapshot.refreshInMs)
    return () => window.clearTimeout(timer)
  }, [snapshot, load])

  useEffect(() => {
    const root = document.documentElement
    if (snapshot == null || snapshot.theme === 'system') {
      delete root.dataset.theme
    } else {
      root.dataset.theme = snapshot.theme
    }
  }, [snapshot])

  if (snapshot == null) {
    return <div className="problem">{problem}</div>
  }
  if (view === 'settings') {
    return <Settings snapshot={snapshot} startAdding={adding} reload={load} onClose={closePanel} />
  }
  if (view === 'about') {
    return <About onClose={closePanel} />
  }
  if (view === 'licence') {
    return <Licence onClose={closePanel} />
  }
  return <Strip snapshot={snapshot} onAddClock={() => openPanel(addClock)} refused={setProblem} />
}
