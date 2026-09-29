import { useEffect, useRef } from 'react'
import { api, type Refused } from './api'

/**
 * The attributes whose change can alter how tall a panel's content is. Style is left out: measuring
 * sets the panel's own, so watching it would measure again at every measurement.
 */
const reshaping = ['class', 'hidden', 'open']

/** naturalHeight answers, in DIP, how tall panel is with room for all its content. */
export function naturalHeight(panel: HTMLElement): number {
  const held = panel.style.height
  panel.style.height = 'auto'
  const height = panel.getBoundingClientRect().height
  panel.style.height = held
  return Math.ceil(height)
}

/**
 * usePanelFit answers a ref for a panel that scrolls its own content. Whenever the height the panel
 * needs to show all of it changes, Go is told; it makes the window that tall where the display has
 * room (FR-621). Anything inside may change it: a clock added, a search answered.
 */
export function usePanelFit<T extends HTMLElement>(refused: Refused) {
  const panel = useRef<T>(null)
  useEffect(() => {
    const element = panel.current
    if (element == null) {
      return
    }
    let reported = 0
    const fit = () => {
      const height = naturalHeight(element)
      if (height > 0 && height !== reported) {
        reported = height
        void api.fitPanel(height, refused)
      }
    }
    fit()
    void document.fonts?.ready.then(fit)
    const watcher = new MutationObserver(fit)
    watcher.observe(element, { childList: true, subtree: true, characterData: true, attributeFilter: reshaping })
    return () => watcher.disconnect()
  }, [refused])
  return panel
}
