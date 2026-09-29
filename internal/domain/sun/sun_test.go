package sun

import (
	"bufio"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

// tolerance is FR-906's: within 0.2 degrees of NOAA's.
const tolerance = 0.2

// FR-906, its acceptance: at eight instants over a year the subsolar point is within 0.2 degrees of
// NOAA's Solar Calculator, read from its own functions and kept in testdata/noaa.tsv.
func TestTheSubsolarPointMatchesNOAA(t *testing.T) {
	t.Parallel()
	file, err := os.Open("testdata/noaa.tsv")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	rows := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		instant, err := time.Parse(time.RFC3339, fields[0])
		if err != nil {
			t.Fatal(err)
		}
		latitude, _ := strconv.ParseFloat(fields[1], 64)
		longitude, _ := strconv.ParseFloat(fields[2], 64)
		got := Subsolar(instant)
		if math.Abs(got.Latitude-latitude) > tolerance || math.Abs(wrapLongitude(got.Longitude-longitude)) > tolerance {
			t.Errorf("%s: got %.4f %.4f, NOAA %.4f %.4f", fields[0], got.Latitude, got.Longitude, latitude, longitude)
		}
		rows++
	}
	if rows != 8 {
		t.Errorf("read %d instants, want 8", rows)
	}
}

// The instant's zone does not move the sun: the same moment written in two zones gives one point.
func TestTheZoneOfTheInstantDoesNotMatter(t *testing.T) {
	t.Parallel()
	utc := time.Date(2026, 6, 21, 9, 15, 0, 0, time.UTC)
	elsewhere := utc.In(time.FixedZone("UTC+10", 10*60*60))
	if Subsolar(utc) != Subsolar(elsewhere) {
		t.Errorf("%+v and %+v", Subsolar(utc), Subsolar(elsewhere))
	}
}

func TestLongitudesWrapIntoOneCircle(t *testing.T) {
	t.Parallel()
	for in, want := range map[float64]float64{180: -180, -180: -180, 190: -170, -190: 170, 540: -180, 0: 0} {
		if got := wrapLongitude(in); math.Abs(got-want) > 1e-9 {
			t.Errorf("%v: got %v, want %v", in, got, want)
		}
	}
}
