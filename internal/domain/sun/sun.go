// Package sun finds where the sun stands overhead at an instant: the subsolar point the sun map is
// lit from (FR-905, FR-906). It reads no clock; the instant is given.
//
// The equations are NOAA's, from the Solar Calculator of its Global Monitoring Laboratory
// (https://gml.noaa.gov/grad/solcalc/), which follow Jean Meeus, Astronomical Algorithms. The
// coefficients are that published series; each group is named for the quantity it gives.
package sun

import (
	"math"
	"time"
)

// Point is a place on the Earth in degrees: latitude north positive, longitude east positive.
type Point struct {
	Latitude, Longitude float64
}

const (
	// j2000 is the Julian day of 2000-01-01 12:00 UTC, the epoch the series is measured from.
	j2000 = 2451545.0
	// unixEpochJulianDay is the Julian day of 1970-01-01 00:00 UTC.
	unixEpochJulianDay = 2440587.5
	daysPerCentury     = 36525.0
	secondsPerDay      = 86400.0
	degreesPerCircle   = 360.0
	halfCircle         = 180.0
	// minutesPerDegree is how many minutes of time the Earth takes to turn one degree.
	minutesPerDegree = 4.0
	minutesPerHour   = 60.0
	// noonMinutes is 12:00 in minutes after midnight.
	noonMinutes = 12 * minutesPerHour
)

// Subsolar answers the point where the sun stands overhead at instant.
func Subsolar(instant time.Time) Point {
	utc := instant.UTC()
	century := (julianDay(utc) - j2000) / daysPerCentury
	obliquity := obliquityCorrected(century)
	declination := degrees(math.Asin(math.Sin(radians(obliquity)) * math.Sin(radians(apparentLongitude(century)))))
	minutes := float64(utc.Hour())*minutesPerHour + float64(utc.Minute()) + float64(utc.Second())/minutesPerHour
	longitude := (noonMinutes - minutes - equationOfTime(century, obliquity)) / minutesPerDegree
	return Point{Latitude: declination, Longitude: wrapLongitude(longitude)}
}

func julianDay(utc time.Time) float64 {
	return float64(utc.UnixNano())/float64(time.Second)/secondsPerDay + unixEpochJulianDay
}

// meanLongitude is the sun's geometric mean longitude, in degrees within a circle.
func meanLongitude(t float64) float64 {
	return math.Mod(280.46646+t*(36000.76983+t*0.0003032), degreesPerCircle)
}

// meanAnomaly is the sun's geometric mean anomaly, in degrees.
func meanAnomaly(t float64) float64 { return 357.52911 + t*(35999.05029-0.0001537*t) }

// eccentricity is the eccentricity of the Earth's orbit.
func eccentricity(t float64) float64 { return 0.016708634 - t*(0.000042037+0.0000001267*t) }

// centre is the sun's equation of the centre, in degrees.
func centre(t float64) float64 {
	m := radians(meanAnomaly(t))
	return math.Sin(m)*(1.914602-t*(0.004817+0.000014*t)) + math.Sin(2*m)*(0.019993-0.000101*t) + math.Sin(3*m)*0.000289
}

// omega is the longitude of the Moon's ascending node, in degrees, which nutation follows.
func omega(t float64) float64 { return 125.04 - 1934.136*t }

// apparentLongitude is the sun's apparent longitude, in degrees.
func apparentLongitude(t float64) float64 {
	return meanLongitude(t) + centre(t) - 0.00569 - 0.00478*math.Sin(radians(omega(t)))
}

// obliquityCorrected is the obliquity of the ecliptic corrected for nutation, in degrees.
func obliquityCorrected(t float64) float64 {
	seconds := 21.448 - t*(46.815+t*(0.00059-t*0.001813))
	mean := 23 + (26+seconds/minutesPerHour)/minutesPerHour
	return mean + 0.00256*math.Cos(radians(omega(t)))
}

// equationOfTime is apparent solar time less mean solar time, in minutes.
func equationOfTime(t, obliquity float64) float64 {
	y := math.Pow(math.Tan(radians(obliquity)/2), 2)
	l0, e, m := radians(meanLongitude(t)), eccentricity(t), radians(meanAnomaly(t))
	value := y*math.Sin(2*l0) - 2*e*math.Sin(m) + 4*e*y*math.Sin(m)*math.Cos(2*l0) -
		0.5*y*y*math.Sin(4*l0) - 1.25*e*e*math.Sin(2*m)
	return degrees(value) * minutesPerDegree
}

// wrapLongitude answers longitude within [-180, 180).
func wrapLongitude(longitude float64) float64 {
	wrapped := math.Mod(longitude+halfCircle, degreesPerCircle)
	if wrapped < 0 {
		wrapped += degreesPerCircle
	}
	return wrapped - halfCircle
}

func radians(deg float64) float64 { return deg * math.Pi / halfCircle }

func degrees(rad float64) float64 { return rad * halfCircle / math.Pi }
