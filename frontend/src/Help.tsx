import { useEffect, useState, type KeyboardEvent, type ReactNode } from 'react'
import { api, type About as AboutFacts } from './api'
import { useAutoScroll } from './autoScroll'
import appIcon from './assets/app-icon.png'

interface PanelProps {
  title: string
  problem: string
  onClose: () => void
  children: ReactNode
}

/**
 * Panel is the frame About and Licence share (FR-607, FR-608): the heading and Close stay put above
 * a body that reads itself when it holds more than fits (FR-609). The panel opens on Close rather
 * than in the body, so opening is never taken for a reader; Escape closes it too.
 */
function Panel({ title, problem, onClose, children }: PanelProps) {
  const reading = useAutoScroll()
  const escape = (event: KeyboardEvent<HTMLElement>) => {
    if (event.key === 'Escape') {
      onClose()
    }
  }
  return (
    <main className="panel" onKeyDown={escape}>
      <header>
        <h1>{title}</h1>
        <button type="button" autoFocus onClick={onClose}>
          Close
        </button>
      </header>
      {problem !== '' && (
        <p className="problem" role="alert">
          {problem}
        </p>
      )}
      <div className="panel-body" ref={reading} tabIndex={0}>
        {children}
      </div>
    </main>
  )
}

/** About names the application, its author and every component it ships, in that order (FR-607). */
export function About({ onClose }: { onClose: () => void }) {
  const [facts, setFacts] = useState<AboutFacts | null>(null)
  const [problem, setProblem] = useState('')
  useEffect(() => {
    void api.about(setProblem).then(setFacts)
  }, [])
  return (
    <Panel title="About" problem={problem} onClose={onClose}>
      <div className="about-head">
        <img src={appIcon} alt="" draggable={false} />
        {facts != null && (
          <>
            <h2>
              {facts.name} {facts.version}
            </h2>
            <p>by {facts.author}</p>
            <p className="muted">{facts.copyright}</p>
          </>
        )}
      </div>
      {facts != null && (
        <>
          <h2>Credits</h2>
          <ul className="credits">
            {facts.credits.map((credit) => (
              <li key={credit.name}>
                <span className="credit-name">{credit.name}</span>, {credit.licence}: {credit.role}
              </li>
            ))}
          </ul>
        </>
      )}
    </Panel>
  )
}

/** Licence shows the whole of the terms the application was built with (FR-608). */
export function Licence({ onClose }: { onClose: () => void }) {
  const [text, setText] = useState('')
  const [problem, setProblem] = useState('')
  useEffect(() => {
    void api.licence(setProblem).then((terms) => setText(terms ?? ''))
  }, [])
  return (
    <Panel title="Licence" problem={problem} onClose={onClose}>
      <pre className="licence-text">{text}</pre>
    </Panel>
  )
}
