import { useCallback, useEffect, useState } from 'react'
import { api, on, type Snapshot } from './api'
import { Settings } from './Settings'
import { Strip } from './Strip'

type View = 'strip' | 'settings'

/** The open-settings event's word for opening straight onto the place search; app.go names it too. */
const addClock = 'add-clock'

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

  const openSettings = useCallback((at?: unknown) => {
    setAdding(at === addClock)
    setView('settings')
    void api.openSettings(setProblem)
  }, [])

  const closeSettings = useCallback(() => {
    setView('strip')
    void api.closeSettings(setProblem).then(load)
  }, [load])

  useEffect(() => {
    load()
    const stopRefresh = on('refresh', load)
    const stopSettings = on('open-settings', openSettings)
    return () => {
      stopRefresh()
      stopSettings()
    }
  }, [load, openSettings])

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
    return <Settings snapshot={snapshot} startAdding={adding} reload={load} onClose={closeSettings} />
  }
  return <Strip snapshot={snapshot} onAddClock={() => openSettings(addClock)} refused={setProblem} />
}
