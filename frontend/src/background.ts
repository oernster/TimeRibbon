/** A colour's red, green and blue channels. */
export type Rgb = [number, number, number]

const opaque = 1

/** The element everything is drawn in; it carries the page's background (FR-622). */
export const rootId = 'root'

const computed = /^rgba?\((\d+),\s*(\d+),\s*(\d+)(?:,\s*([\d.]+))?\)$/

/** rgbOf reads a computed CSS colour as its channels; null when it is not an opaque rgb() or rgba(). */
export function rgbOf(css: string): Rgb | null {
  const match = computed.exec(css.trim())
  if (match == null || (match[4] !== undefined && Number(match[4]) !== opaque)) {
    return null
  }
  return [Number(match[1]), Number(match[2]), Number(match[3])]
}

/**
 * backgroundReporter hands report the colour the page paints behind everything, so Go can give the
 * window that colour rather than white while the page catches up with a new size. check reports it
 * when it has changed since the last report; the system switching light or dark checks by itself,
 * since the system theme changes the colour without any choice changing. stop ends that listening.
 */
export function backgroundReporter(report: (...rgb: Rgb) => void, view: Window = window) {
  let last = ''
  const check = () => {
    const painted = view.document.getElementById(rootId) ?? view.document.body
    const css = view.getComputedStyle(painted).backgroundColor
    const rgb = rgbOf(css)
    if (rgb == null || css === last) {
      return
    }
    last = css
    report(...rgb)
  }
  const scheme = view.matchMedia?.('(prefers-color-scheme: dark)')
  scheme?.addEventListener('change', check)
  return { check, stop: () => scheme?.removeEventListener('change', check) }
}
