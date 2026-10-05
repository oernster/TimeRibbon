import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { App } from './App'
import { rootId } from '@oernster/ribbonkit'
import '@oernster/ribbonkit/theme.css'
import '@oernster/ribbonkit/colours.css'
import './dials.css'
import '@oernster/ribbonkit/controls.css'
import './app.css'
import './settings.css'
import './help.css'

const root = document.getElementById(rootId)
if (root != null) {
  createRoot(root).render(
    <StrictMode>
      <App />
    </StrictMode>,
  )
}
