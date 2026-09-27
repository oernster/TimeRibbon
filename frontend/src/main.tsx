import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { App } from './App'
import './theme.css'
import './app.css'
import './settings.css'
import './help.css'

const root = document.getElementById('root')
if (root != null) {
  createRoot(root).render(
    <StrictMode>
      <App />
    </StrictMode>,
  )
}
