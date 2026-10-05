package structural

// ribbonkit's half of the page leaves this repository with the rest of the kit (WeatherRibbon
// CON-10), so it may reach nothing of TimeRibbon's. TestTheKitImportsNothingOfTimeRibbon holds the
// Go half to that; this holds the page half: every relative import and every url() in ribbonkit/web
// stays inside the kit (the self-reading cycle it imports lives beside the setup page, which can
// import nothing). Every package it names is one the kit's package.json states.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// kitTestTools are the packages the kit's suites and its testing stand-in run under, with why each
// is not among the kit's peer dependencies: an application's own test set-up supplies them.
var kitTestTools = map[string]string{
	"vitest":                 "the test runner every suite is written for",
	"@testing-library/react": "renders the kit's components in its suites",
}

var (
	// pageSpecifier finds a script import's module: from 'x', import 'x' and import('x').
	pageSpecifier = regexp.MustCompile(`(?:\bfrom|\bimport)\s*\(?\s*['"]([^'"]+)['"]`)
	// cssReference finds what a stylesheet reaches: url(x) and @import 'x'.
	cssReference = regexp.MustCompile(`url\(\s*['"]?([^'")]+)|@import\s+['"]([^'"]+)['"]`)
	// pageComment is a block comment or a whole-line one, whose words name modules without importing
	// them (index.ts says how an application imports the kit).
	pageComment = regexp.MustCompile(`(?s)/\*.*?\*/|(?m)^\s*//.*$`)
)

// kitPeers answers the packages ribbonkit/package.json names as its peer dependencies.
func kitPeers(t *testing.T) map[string]bool {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(repoRoot(t), kitTree, "package.json"))
	if err != nil {
		t.Fatal(err)
	}
	var stated struct {
		PeerDependencies map[string]string `json:"peerDependencies"`
	}
	if err := json.Unmarshal(raw, &stated); err != nil {
		t.Fatal(err)
	}
	peers := map[string]bool{}
	for name := range stated.PeerDependencies {
		peers[name] = true
	}
	if len(peers) == 0 {
		t.Fatal("ribbonkit/package.json states no peer dependencies, the read is wrong")
	}
	return peers
}

// packageOf answers the package a bare specifier names: its first part; two for a scoped one.
func packageOf(specifier string) string {
	parts := strings.Split(specifier, "/")
	if strings.HasPrefix(specifier, "@") && len(parts) > 1 {
		return parts[0] + "/" + parts[1]
	}
	return parts[0]
}

// referencesOf answers every module or file path reaches, by the form its extension is written in.
func referencesOf(t *testing.T, path string) []string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	form := pageSpecifier
	if filepath.Ext(path) == ".css" {
		form = cssReference
	}
	var found []string
	for _, match := range form.FindAllStringSubmatch(pageComment.ReplaceAllString(string(raw), ""), -1) {
		for _, group := range match[1:] {
			if group != "" {
				found = append(found, strings.TrimSpace(group))
			}
		}
	}
	return found
}

func TestTheKitPageReachesNothingOfTimeRibbon(t *testing.T) {
	root := repoRoot(t)
	kit := filepath.Join(root, kitTree)
	web := filepath.Join(kit, "web")
	peers := kitPeers(t)
	read := 0
	for _, path := range frontendFiles(t) {
		inside, err := filepath.Rel(web, path)
		if err != nil || strings.HasPrefix(inside, "..") {
			continue
		}
		read++
		for _, reference := range referencesOf(t, path) {
			if strings.HasPrefix(reference, ".") {
				target, _ := filepath.Rel(kit, filepath.Join(filepath.Dir(path), reference))
				if strings.HasPrefix(target, "..") {
					t.Errorf("%s reaches %s, outside the kit", filepath.ToSlash(inside), reference)
				}
				continue
			}
			name := packageOf(reference)
			if _, tool := kitTestTools[name]; !peers[name] && !tool {
				t.Errorf("%s imports %s, which ribbonkit/package.json does not depend on", filepath.ToSlash(inside), reference)
			}
		}
	}
	if read == 0 {
		t.Fatal("no file of ribbonkit/web was read, the walk is wrong")
	}
}
