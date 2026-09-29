import { useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react'
import type { Refused, Size, SunMap as SunMapWire } from './api'
import dayArt from './assets/sun-day.jpg'
import nightArt from './assets/sun-night.jpg'
import { showsTheMenu, useDrag } from './drag'
import { placeLabels, spots, type Spot } from './labels'
import { blend, project } from './sunLight'

interface Props {
  sunMap: SunMapWire
  /** The map's size in CSS pixels; the picture is drawn at the display's own resolution. */
  width: number
  height: number
  dragThreshold: Size
  refused: Refused
}

/** The two pictures, named for the notice raised when one cannot be read (FR-911). */
const pictures = { day: { art: dayArt, name: 'the day map' }, night: { art: nightArt, name: 'the night map' } }

/** loaded answers a picture once it has loaded; a rejection names it. */
function loaded(picture: { art: string; name: string }): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const image = new Image()
    image.onload = () => resolve(image)
    image.onerror = () => reject(new Error(`${picture.name} could not be read`))
    image.src = picture.art
  })
}

/** styleLength answers a length token from the page's style in CSS pixels; 0 where it is not set. */
function styleLength(style: CSSStyleDeclaration, token: string): number {
  return Number.parseFloat(style.getPropertyValue(token)) || 0
}

/** pixelsOf answers a picture scaled to width by height as RGBA pixels. */
function pixelsOf(image: HTMLImageElement, width: number, height: number): Uint8ClampedArray {
  const canvas = document.createElement('canvas')
  canvas.width = width
  canvas.height = height
  const context = canvas.getContext('2d')
  if (context == null) {
    throw new Error('the map could not be drawn: no canvas')
  }
  context.drawImage(image, 0, 0, width, height)
  return context.getImageData(0, 0, width, height).data
}

/**
 * SunMap is the world lit where it is day and dark with city lights where it is night, the clocks'
 * places marked (FR-905 to FR-908). It is drawn again for each snapshot's subsolar point (FR-907). A
 * press on it drags the ribbon with it (FR-909). A picture that cannot be read leaves words in its
 * place rather than a blank map (FR-911). Each label is measured once drawn, then stood clear of the
 * other labels and dots (FR-914).
 */
export function SunMap({ sunMap, width, height, dragThreshold, refused }: Props) {
  const canvas = useRef<HTMLCanvasElement>(null)
  const root = useRef<HTMLDivElement>(null)
  const [images, setImages] = useState<[HTMLImageElement, HTMLImageElement] | null>(null)
  const [problem, setProblem] = useState('')
  const [placed, setPlaced] = useState<Spot[]>([])
  const drag = useDrag(dragThreshold)
  const dots = useMemo(() => sunMap.marks.map((mark) => project(mark.latitude, mark.longitude, width, height)), [sunMap.marks, width, height])

  useLayoutEffect(() => {
    const element = root.current
    if (element == null) {
      return
    }
    const style = getComputedStyle(element)
    const sizes = Array.from(element.querySelectorAll<HTMLElement>('.mark-label'), (label) => ({ width: label.offsetWidth, height: label.offsetHeight }))
    const next = placeLabels(dots, sizes, { width, height }, styleLength(style, '--mark-size'), styleLength(style, '--mark-gap'))
    setPlaced((last) => (last.length === next.length && last.every((spot, index) => spot === next[index]) ? last : next))
  }, [dots, width, height])

  useEffect(() => {
    Promise.all([loaded(pictures.day), loaded(pictures.night)]).then(setImages, (failure: Error) => setProblem(`The sun map cannot be shown: ${failure.message}.`))
  }, [])

  useEffect(() => {
    const target = canvas.current
    if (images == null || target == null) {
      return
    }
    const pixelWidth = Math.max(1, Math.round(width * window.devicePixelRatio))
    const pixelHeight = Math.max(1, Math.round(height * window.devicePixelRatio))
    target.width = pixelWidth
    target.height = pixelHeight
    const context = target.getContext('2d')
    if (context == null) {
      setProblem('The sun map cannot be shown: the page has no canvas.')
      return
    }
    const out = context.createImageData(pixelWidth, pixelHeight)
    blend(pixelsOf(images[0], pixelWidth, pixelHeight), pixelsOf(images[1], pixelWidth, pixelHeight), out.data, pixelWidth, pixelHeight, sunMap.latitude, sunMap.longitude)
    context.putImageData(out, 0, 0)
  }, [images, width, height, sunMap.latitude, sunMap.longitude])

  return (
    <div ref={root} className="sun-map" style={{ width, height }} {...drag} onContextMenu={showsTheMenu(refused)}>
      {problem !== '' ? (
        <div className="problem" role="alert">{problem}</div>
      ) : (
        <canvas ref={canvas} style={{ width, height }} />
      )}
      {sunMap.marks.map((mark, index) => (
        <div key={mark.label + mark.latitude + mark.longitude} className="mark" style={{ left: dots[index].x, top: dots[index].y }}>
          <span className="dot" />
          <span className={`mark-label ${placed[index] ?? spots[0]}`}>{mark.label}</span>
        </div>
      ))}
    </div>
  )
}
