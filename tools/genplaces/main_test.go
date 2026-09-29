package main

import (
	"math"
	"slices"
	"strings"
	"testing"
)

func TestPlacesJoinsCountriesAndSortsByZone(t *testing.T) {
	t.Parallel()
	zones, err := parseTable(strings.NewReader("# comment\n\nNO\t+5955+01045\tEurope/Oslo\nAE,OM\t+2518+05518\tAsia/Dubai\tGulf\n"), zoneMinColumns)
	if err != nil {
		t.Fatal(err)
	}
	countries := [][]string{{"NO", "Norway"}, {"AE", "United Arab Emirates"}, {"OM", "Oman"}}
	got, err := places(zones, countries)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"Asia/Dubai\tUnited Arab Emirates; Oman\t25.3000\t55.3000", "Europe/Oslo\tNorway\t59.9167\t10.7500"}
	if !slices.Equal(got, want) {
		t.Errorf("got %q", got)
	}
}

func TestAnUnknownCountryCodeOrAShortLineIsRefused(t *testing.T) {
	t.Parallel()
	if _, err := places([][]string{{"ZZ", "+5955+01045", "Nowhere/Town"}}, nil); err == nil {
		t.Error("an unknown code was accepted")
	}
	if _, err := parseTable(strings.NewReader("NO\tEurope/Oslo\n"), zoneMinColumns); err == nil {
		t.Error("a short line was accepted")
	}
}

// FR-908: zone.tab's ISO 6709 coordinates, with and without seconds, west and south negative.
func TestCoordinatesAreReadAsDecimalDegrees(t *testing.T) {
	t.Parallel()
	cases := map[string][2]float64{
		"+513030-0000731": {51.508333, -0.125278},
		"-3352+15113":     {-33.866667, 151.216667},
		"+4734-05243":     {47.566667, -52.716667},
	}
	for text, want := range cases {
		latitude, longitude, err := coordinates(text)
		if err != nil || math.Abs(latitude-want[0]) > 1e-5 || math.Abs(longitude-want[1]) > 1e-5 {
			t.Errorf("%s: got %f %f %v, want %f %f", text, latitude, longitude, err, want[0], want[1])
		}
	}
}

func TestMalformedCoordinatesAreRefused(t *testing.T) {
	t.Parallel()
	for _, text := range []string{"", "+5955", "5955+01045", "+595+01045", "+59x5+01045", "+5955*01045"} {
		if _, _, err := coordinates(text); err == nil {
			t.Errorf("%q was accepted", text)
		}
	}
	if _, err := places([][]string{{"NO", "bad", "Europe/Oslo"}}, [][]string{{"NO", "Norway"}}); err == nil {
		t.Error("a zone with bad coordinates was accepted")
	}
}
