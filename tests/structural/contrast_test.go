package structural

// Label, time, date and zone mark text meets a contrast ratio of at least 4.5:1 in both themes, on
// every colour scheme the menus offer (NFR-U-1, FR-606, FR-611).
//
// The colours are read from the very files the page loads: theme.css holds Classic, light in :root
// and dark under [data-theme='dark']; colours.css holds the other schemes, each token written as
// light-dark(light, dark). A scheme that leaves a token out draws it in Classic's value for the same
// theme, so that is the value checked. The label, time and date draw in --text and the zone mark in
// --text-muted (app.css); --problem is checked too, since colours.css states it meets the same floor
// and colours_test.go lets schemes inherit it on that promise. The cell paints no background of its
// own, so its text lies on #root's --surface; both --cell and --surface are checked, as colours.css
// states. The test lives here rather than in Vitest because Vitest hands a CSS import back empty.

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"testing"

	"github.com/oernster/timeribbon/internal/domain/settings"
)

// minContrast is WCAG 2.x's AA floor for body text (success criterion 1.4.3).
const minContrast = 4.5

// textTokens are the colours the ribbon's words are drawn in; backgrounds are what they lie on.
var (
	textTokens  = []string{"text", "text-muted", "problem"}
	backgrounds = []string{"cell", "surface"}
)

// WCAG 2.x relative luminance: channels are linearised below linearLimit and by the sRGB curve
// above it, then weighted by the eye's sensitivity to each. flare is the ambient light WCAG adds to
// both luminances before dividing them.
const (
	channelMax    = 255
	linearLimit   = 0.04045
	linearDivisor = 12.92
	curveOffset   = 0.055
	curveDivisor  = 1.055
	curvePower    = 2.4
	flare         = 0.05
)

var luminanceWeights = [3]float64{0.2126, 0.7152, 0.0722}

type theme string

const (
	lightTheme theme = "light"
	darkTheme  theme = "dark"
)

var (
	darkBlock   = regexp.MustCompile(`(?m)^:root\[data-theme='dark'\] \{([^}]*)\}`)
	systemDark  = regexp.MustCompile(`@media \(prefers-color-scheme: dark\) \{\s*:root:not\(\[data-theme='light'\]\) \{([^}]*)\}`)
	declaration = regexp.MustCompile(`--([\w-]+):\s*([^;]+);`)
	lightDark   = regexp.MustCompile(`^light-dark\(\s*([^,]+?)\s*,\s*([^)]+?)\s*\)$`)
	shortHex    = regexp.MustCompile(`^#([0-9a-fA-F])([0-9a-fA-F])([0-9a-fA-F])$`)
	longHex     = regexp.MustCompile(`^#([0-9a-fA-F]{2})([0-9a-fA-F]{2})([0-9a-fA-F]{2})$`)
)

func declarations(body string) map[string]string {
	tokens := map[string]string{}
	for _, match := range declaration.FindAllStringSubmatch(body, -1) {
		tokens[match[1]] = match[2]
	}
	return tokens
}

// channels reads #rgb or #rrggbb, refusing any other form rather than skipping the check.
func channels(colour string) ([3]float64, error) {
	var digits []string
	if match := shortHex.FindStringSubmatch(colour); match != nil {
		for _, digit := range match[1:] {
			digits = append(digits, digit+digit)
		}
	} else if match := longHex.FindStringSubmatch(colour); match != nil {
		digits = match[1:]
	} else {
		return [3]float64{}, fmt.Errorf("cannot read the colour %q; teach contrast_test.go its form", colour)
	}
	var rgb [3]float64
	for index, pair := range digits {
		value, err := strconv.ParseUint(pair, 16, 8)
		if err != nil {
			return rgb, err
		}
		rgb[index] = float64(value)
	}
	return rgb, nil
}

func luminance(colour string) (float64, error) {
	rgb, err := channels(colour)
	if err != nil {
		return 0, err
	}
	sum := 0.0
	for index, channel := range rgb {
		unit := channel / channelMax
		linear := unit / linearDivisor
		if unit > linearLimit {
			linear = math.Pow((unit+curveOffset)/curveDivisor, curvePower)
		}
		sum += linear * luminanceWeights[index]
	}
	return sum, nil
}

// contrast answers WCAG 2.x's contrast ratio between two colours, whichever is lighter.
func contrast(first, second string) (float64, error) {
	a, err := luminance(first)
	if err != nil {
		return 0, err
	}
	b, err := luminance(second)
	if err != nil {
		return 0, err
	}
	return (math.Max(a, b) + flare) / (math.Min(a, b) + flare), nil
}

// palette answers the colour every token has in one scheme and theme.
type palette func(token string) (string, error)

// palettes answers each offered scheme's palette per theme, a scheme's missing token falling back
// to Classic's value for the same theme as the cascade does.
func palettes(t *testing.T) map[settings.Colour]map[theme]palette {
	t.Helper()
	themeCss := readFrontend(t, "theme.css")
	light := classicBlock.FindStringSubmatch(themeCss)
	dark := darkBlock.FindStringSubmatch(themeCss)
	if light == nil || dark == nil {
		t.Fatal("theme.css lacks its :root or its [data-theme='dark'] block")
	}
	classic := map[theme]map[string]string{lightTheme: declarations(light[1]), darkTheme: declarations(dark[1])}
	stated := map[string]map[string]string{}
	for _, match := range schemeBlock.FindAllStringSubmatch(readFrontend(t, "colours.css"), -1) {
		stated[match[1]] = declarations(match[2])
	}
	all := map[settings.Colour]map[theme]palette{}
	for _, colour := range settings.Colours {
		own := stated[string(colour)]
		all[colour] = map[theme]palette{}
		for _, side := range []theme{lightTheme, darkTheme} {
			all[colour][side] = func(token string) (string, error) {
				value, found := own[token]
				if !found {
					value, found = classic[side][token]
				}
				if !found {
					return "", fmt.Errorf("--%s is stated by neither %s nor Classic", token, colour)
				}
				if pair := lightDark.FindStringSubmatch(value); pair != nil {
					if side == lightTheme {
						return pair[1], nil
					}
					return pair[2], nil
				}
				return value, nil
			}
		}
	}
	return all
}

func TestContrastIsComputedAsTheStandardStatesIt(t *testing.T) {
	// Black on white is the scale's top, 21:1; a colour on itself is its bottom, 1:1. The two greys
	// are the ones usually quoted either side of the floor, which only the sRGB curve gets right.
	cases := []struct {
		first, second string
		want          float64
	}{
		{"#000000", "#ffffff", 21},
		{"#fff", "#000", 21},
		{"#777777", "#777777", 1},
		{"#767676", "#ffffff", 4.54},
		{"#777777", "#ffffff", 4.48},
	}
	// tolerance is half the last place of a ratio quoted to two decimals.
	const tolerance = 0.005
	for _, c := range cases {
		got, err := contrast(c.first, c.second)
		if err != nil || math.Abs(got-c.want) > tolerance {
			t.Errorf("contrast(%s, %s) = %v, %v; want %v", c.first, c.second, got, err, c.want)
		}
	}
	if _, err := contrast("rgb(0 0 0)", "#fff"); err == nil {
		t.Error("a colour form the test cannot read was accepted")
	}
}

func TestClassicDarkIsTheSameUnderTheSystemAsWhenChosen(t *testing.T) {
	// The test reads the chosen dark block; the page shows the system one under System (FR-606).
	themeCss := readFrontend(t, "theme.css")
	system := systemDark.FindStringSubmatch(themeCss)
	chosen := darkBlock.FindStringSubmatch(themeCss)
	if system == nil || chosen == nil {
		t.Fatal("theme.css lacks a dark block")
	}
	if fmt.Sprint(declarations(system[1])) != fmt.Sprint(declarations(chosen[1])) {
		t.Errorf("the system's dark block differs from the chosen one:\n%v\n%v", declarations(system[1]), declarations(chosen[1]))
	}
}

func TestTextMeetsTheContrastFloorOnEverySchemeAndTheme(t *testing.T) {
	for colour, sides := range palettes(t) {
		for side, colourOf := range sides {
			for _, token := range textTokens {
				for _, background := range backgrounds {
					name := fmt.Sprintf("%s %s --%s on --%s", colour, side, token, background)
					fore, err := colourOf(token)
					if err != nil {
						t.Errorf("%s: %v", name, err)
						continue
					}
					back, err := colourOf(background)
					if err != nil {
						t.Errorf("%s: %v", name, err)
						continue
					}
					ratio, err := contrast(fore, back)
					if err != nil {
						t.Errorf("%s: %v", name, err)
						continue
					}
					if ratio < minContrast {
						t.Errorf("%s: %s on %s is %.2f:1, under %.1f:1", name, fore, back, ratio, minContrast)
					}
				}
			}
		}
	}
}
