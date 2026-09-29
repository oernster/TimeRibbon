import { describe, expect, it } from 'vitest'
import { DUSK, blend, dayWeight, latitudeAt, longitudeAt, project, solarAltitude } from './sunLight'

describe('sun light (FR-905)', () => {
  it('puts the sun overhead at the subsolar point and on the horizon a quarter turn away', () => {
    expect(solarAltitude(0, 0, 0, 0)).toBeCloseTo(90)
    expect(solarAltitude(0, 90, 0, 0)).toBeCloseTo(0)
    expect(solarAltitude(0, 180, 0, 0)).toBeCloseTo(-90)
    expect(solarAltitude(51.5, 0, 23.44, 0)).toBeCloseTo(90 - 51.5 + 23.44)
  })

  it('shows all day above the horizon, all night below dusk and half at half dusk', () => {
    expect(dayWeight(10)).toBe(1)
    expect(dayWeight(0)).toBe(1)
    expect(dayWeight(DUSK)).toBe(0)
    expect(dayWeight(-30)).toBe(0)
    expect(dayWeight(DUSK / 2)).toBeCloseTo(0.5)
  })

  it('maps pixels to places and places to pixels on an equirectangular map', () => {
    expect(longitudeAt(0, 360)).toBeCloseTo(-179.5)
    expect(latitudeAt(0, 180)).toBeCloseTo(89.5)
    expect(project(0, 0, 960, 480)).toEqual({ x: 480, y: 240 })
    expect(project(90, -180, 960, 480)).toEqual({ x: 0, y: 0 })
  })

  it('draws day under the sun, night opposite it and a mix at half dusk', () => {
    const width = 4
    const height = 2
    const size = width * height * 4
    const day = new Uint8ClampedArray(size).fill(200)
    const night = new Uint8ClampedArray(size).fill(0)
    const out = new Uint8ClampedArray(size)
    // The sun over longitude -135, the centre of the first column; the third column is its night.
    blend(day, night, out, width, height, 0, -135)
    expect(out[0]).toBe(200)
    expect(out[2 * 4]).toBe(0)
    // One row lies on the equator, where a sun 96 degrees round from the first column's centre
    // stands 6 degrees below its horizon: half dusk, so half of each picture.
    const row = width * 4
    const twilight = new Uint8ClampedArray(row)
    blend(day.subarray(0, row), night.subarray(0, row), twilight, width, 1, 0, -135 + 90 + 6)
    expect(twilight[0]).toBe(100)
  })
})
