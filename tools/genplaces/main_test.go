package main

import (
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
	want := []string{"Asia/Dubai\tUnited Arab Emirates; Oman", "Europe/Oslo\tNorway"}
	if !slices.Equal(got, want) {
		t.Errorf("got %q", got)
	}
}

func TestAnUnknownCountryCodeOrAShortLineIsRefused(t *testing.T) {
	t.Parallel()
	if _, err := places([][]string{{"ZZ", "", "Nowhere/Town"}}, nil); err == nil {
		t.Error("an unknown code was accepted")
	}
	if _, err := parseTable(strings.NewReader("NO\tEurope/Oslo\n"), zoneMinColumns); err == nil {
		t.Error("a short line was accepted")
	}
}
