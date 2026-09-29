import { useEffect, useRef, useState } from 'react'
import type { Size, SunMap as SunMapWire } from './api'
import dayArt from './assets/sun-day.jpg'
import nightArt from './assets/sun-night.jpg'
import { useDrag } from './drag'
import { blend, project } from './sunLight'

interface Props {
  sunMap: SunMapWire
  /** The map's size in CSS pixels; the picture is drawn at the display's own resolution. */
  width: number
  height: number
  dragThreshold: Size
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
 * place rather than a blank map (FR-911).
 */
export function SunMap({ sunMap, width, height, dragThreshold }: Props) {
  const canvas = useRef<HTMLCanvasElement>(null)
  const [images, setImages] = useState<[HTMLImageElement, HTMLImageElement] | null>(null)
  const [problem, setProblem] = useState('')
  const drag = useDrag(dragThreshold)

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
    <div className="sun-map" style={{ width, height }} {...drag}>
      {problem !== '' ? (
        <div className="problem" role="alert">{problem}</div>
      ) : (
        <canvas ref={canvas} style={{ width, height }} />
      )}
      {sunMap.marks.map((mark) => {
        const at = project(mark.latitude, mark.longitude, width, height)
        return (
          <div key={mark.label + mark.latitude + mark.longitude} className="mark" style={{ left: at.x, top: at.y }}>
            <span className="dot" />
            <span className="mark-label">{mark.label}</span>
          </div>
        )
      })}
    </div>
  )
}
