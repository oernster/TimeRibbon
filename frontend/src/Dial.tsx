// An analogue dial (FR-603): twelve ticks, an hour hand and a minute hand, no second hand. Drawn in
// a 100 by 100 box centred on 50, 50.

const centre = 50
const faceRadius = 46
const tickOuter = 42
const hourTickInner = 35
const hourHandLength = 24
const minuteHandLength = 36
const hours = 12
const degreesPerHour = 360 / hours

interface Props {
  hourAngle: number
  minuteAngle: number
}

function Hand({ angle, length, width }: { angle: number; length: number; width: number }) {
  return (
    <line
      x1={centre}
      y1={centre}
      x2={centre}
      y2={centre - length}
      stroke="var(--dial-hand)"
      strokeWidth={width}
      strokeLinecap="round"
      transform={`rotate(${angle} ${centre} ${centre})`}
    />
  )
}

export function Dial({ hourAngle, minuteAngle }: Props) {
  return (
    <svg className="dial" viewBox="0 0 100 100" aria-hidden="true">
      <circle cx={centre} cy={centre} r={faceRadius} fill="var(--dial-face)" stroke="var(--divider)" strokeWidth={2} />
      {Array.from({ length: hours }, (_, hour) => (
        <line
          key={hour}
          x1={centre}
          y1={centre - tickOuter}
          x2={centre}
          y2={centre - hourTickInner}
          stroke="var(--dial-tick)"
          strokeWidth={hour % 3 === 0 ? 3 : 1.5}
          transform={`rotate(${hour * degreesPerHour} ${centre} ${centre})`}
        />
      ))}
      <Hand angle={hourAngle} length={hourHandLength} width={5} />
      <Hand angle={minuteAngle} length={minuteHandLength} width={3} />
      <circle cx={centre} cy={centre} r={3} fill="var(--dial-hand)" />
    </svg>
  )
}
