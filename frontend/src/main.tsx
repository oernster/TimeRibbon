import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { App } from './App'
import { rootId } from '@oernster/ribbonkit'
import './theme.css'
import './colours.css'
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
