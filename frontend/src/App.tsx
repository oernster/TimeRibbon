import { useCallback, useEffect, useState } from 'react'
import { api, on, type Snapshot, type UpdateStatus } from './api'
import { About, Licence, Update } from './Help'
import { watchPixelRatio } from './pixelRatio'
import { scrollbarThickness } from './scrollbar'
import { Settings } from './Settings'
import { Ribbon } from './Ribbon'

/** The panels the window can become (CON-6); app.go names each in its open-panel event. */
type Panel = 'settings' | 'about' | 'licence' | 'update'
type View = 'ribbon' | Panel

/**
 * The open-panel event's words, each naming the panel it opens; add-clock opens Settings on the place
 * search; update carries the check's outcome with it. The keys are quoted so the structural test
 * can find each word app.go sends.
 */
const addClock = 'add-clock'
const panelFor: Record<string, Panel> = {
  'settings': 'settings', [addClock]: 'settings', 'about': 'about', 'licence': 'licence', 'update': 'update',
}

/**
 * App holds the snapshot and which surface the window shows. The snapshot is taken again at each
 * minute boundary Go names (FR-208) and whenever Go says the time, the displays or a choice changed
 * (FR-209).
 */
export function App() {
  const [snapshot, setSnapshot] = useState<Snapshot | null>(null)
  const [problem, setProblem] = useState('')
  const [view, setView] = useState<View>('ribbon')
  const [adding, setAdding] = useState(false)
  const [update, setUpdate] = useState<UpdateStatus | null>(null)

  const load = useCallback(() => {
    void api.snapshot(setProblem).then((next) => {
      if (next != null) {
        setSnapshot(next)
        setProblem('')
      }
    })
  }, [])

  const openPanel = useCallback((at?: unknown, outcome?: unknown) => {
    setAdding(at === addClock)
    setUpdate((outcome as UpdateStatus | undefined) ?? null)
    setView(panelFor[String(at)] ?? 'settings')
    void api.openPanel(setProblem)
  }, [])

  const closePanel = useCallback(() => {
    setView('ribbon')
    void api.closePanel(setProblem).then(load)
  }, [load])

  useEffect(() => {
    // Go makes room for the scroll bar a scrolling ribbon shows, which only the page can measure.
    void api.setScrollbar(scrollbarThickness(), setProblem)
  }, [])

  // Go sizes the window by the scale the page is really drawn at, which only the page knows.
  useEffect(() => watchPixelRatio((ratio) => void api.setPixelRatio(ratio, setProblem)), [])

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
    // The colour scheme (FR-611); colours.css keys its schemes off it, Classic being theme.css's own.
    root.dataset.colour = snapshot?.colour ?? 'classic'
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
  if (view === 'update' && update != null) {
    return <Update status={update} onClose={closePanel} />
  }
  return <Ribbon snapshot={snapshot} onAddClock={() => openPanel(addClock)} refused={setProblem} />
}
