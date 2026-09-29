/**
 * The sun map's light (FR-905): how high the sun stands at each point of a whole-world map and how
 * much of the day picture shows there. The map is equirectangular: longitude -180 to 180 left to
 * right, latitude 90 to -90 top to bottom.
 */

/** Below this solar altitude, in degrees, only the night picture shows: nautical dusk (FR-905). */
export const DUSK = -12

const DEGREES = Math.PI / 180
const HALF_TURN = 180
const QUARTER_TURN = 90
const CHANNELS = 4

/** solarAltitude answers the sun's height in degrees above the horizon at a place, from the subsolar point. */
export function solarAltitude(latitude: number, longitude: number, sunLatitude: number, sunLongitude: number): number {
  const sine =
    Math.sin(latitude * DEGREES) * Math.sin(sunLatitude * DEGREES) +
    Math.cos(latitude * DEGREES) * Math.cos(sunLatitude * DEGREES) * Math.cos((longitude - sunLongitude) * DEGREES)
  return Math.asin(Math.max(-1, Math.min(1, sine))) / DEGREES
}

/** dayWeight answers how much of the day picture shows at a solar altitude: all above 0, none below DUSK, in proportion between. */
export function dayWeight(altitude: number): number {
  if (altitude >= 0) {
    return 1
  }
  if (altitude <= DUSK) {
    return 0
  }
  return (altitude - DUSK) / -DUSK
}

/** longitudeAt and latitudeAt answer the place at the centre of a map pixel. */
export function longitudeAt(x: number, width: number): number {
  return -HALF_TURN + ((x + 0.5) / width) * 2 * HALF_TURN
}

export function latitudeAt(y: number, height: number): number {
  return QUARTER_TURN - ((y + 0.5) / height) * HALF_TURN
}

/** project answers where a place falls on a map of width by height. */
export function project(latitude: number, longitude: number, width: number, height: number): { x: number; y: number } {
  return {
    x: ((longitude + HALF_TURN) / (2 * HALF_TURN)) * width,
    y: ((QUARTER_TURN - latitude) / HALF_TURN) * height,
  }
}

/**
 * blend writes into out, pixel by pixel, the day picture where the sun is up and the night picture
 * with its lights where it is down, mixed through twilight (FR-905). day, night and out are RGBA
 * pixels of the same width and height. The terms that depend on the row alone or the column alone
 * are worked out once each rather than per pixel.
 */
export function blend(
  day: Uint8ClampedArray, night: Uint8ClampedArray, out: Uint8ClampedArray,
  width: number, height: number, sunLatitude: number, sunLongitude: number,
): void {
  const sinSun = Math.sin(sunLatitude * DEGREES)
  const cosSun = Math.cos(sunLatitude * DEGREES)
  const cosHour = new Float64Array(width)
  for (let x = 0; x < width; x++) {
    cosHour[x] = Math.cos((longitudeAt(x, width) - sunLongitude) * DEGREES)
  }
  for (let y = 0; y < height; y++) {
    const latitude = latitudeAt(y, height) * DEGREES
    const rowSin = Math.sin(latitude) * sinSun
    const rowCos = Math.cos(latitude) * cosSun
    for (let x = 0; x < width; x++) {
      const altitude = Math.asin(Math.max(-1, Math.min(1, rowSin + rowCos * cosHour[x]))) / DEGREES
      const weight = dayWeight(altitude)
      const at = (y * width + x) * CHANNELS
      for (let channel = 0; channel < CHANNELS; channel++) {
        out[at + channel] = day[at + channel] * weight + night[at + channel] * (1 - weight)
      }
    }
  }
}
