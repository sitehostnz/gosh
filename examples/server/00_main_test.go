package main

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// stepLineRE matches a row of a numbered step map, in either the
// package doc comment or the README's fenced block.
var stepLineRE = regexp.MustCompile(`^(?://\t|)\s*(\d{2})\s{2,}\S`)

// mapNumbers reads the step numbers out of a numbered map in text.
func mapNumbers(t *testing.T, body, startAfter, endBefore string) []int {
	t.Helper()

	i := strings.Index(body, startAfter)
	if i < 0 {
		t.Fatalf("could not find %q to start the map", startAfter)
	}
	rest := body[i+len(startAfter):]
	if j := strings.Index(rest, endBefore); j >= 0 {
		rest = rest[:j]
	}

	seen := map[int]bool{}
	for _, line := range strings.Split(rest, "\n") {
		m := stepLineRE.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		var n int
		if _, err := fmt.Sscanf(m[1], "%d", &n); err == nil {
			seen[n] = true
		}
	}
	out := make([]int, 0, len(seen))
	for n := range seen {
		out = append(out, n)
	}
	sort.Ints(out)
	return out
}

// registeredNumbers lists the distinct step orders steps() registers.
func registeredNumbers() []int {
	seen := map[int]bool{}
	for _, s := range steps() {
		seen[s.order] = true
	}
	out := make([]int, 0, len(seen))
	for n := range seen {
		out = append(out, n)
	}
	sort.Ints(out)
	return out
}

// TestStepMapsMatchTheSteps stops the documentation drifting from the
// program a third time.
//
// Both the package doc comment and the README carry a hand-maintained
// map of the numbered steps, and both are the first thing a reader
// meets — one on pkg.go.dev, the other in the repository. Twice now a
// round of changes has added steps without updating them, and the
// second time three of the missing steps wrote to real infrastructure.
//
// It compares numbers rather than descriptions on purpose: the wording
// is meant to differ between a godoc table and a README, but a number
// that exists in one and not the other is always a mistake.
func TestStepMapsMatchTheSteps(t *testing.T) {
	t.Parallel()

	want := registeredNumbers()

	doc, err := os.ReadFile("00_main.go")
	if err != nil {
		t.Fatalf("reading the package doc: %v", err)
	}
	got := mapNumbers(t, string(doc), "numbered by the order they must run", "// The 50/60/70 ordering")
	assertSameNumbers(t, "the package doc comment in 00_main.go", got, want)

	readme, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("reading the README: %v", err)
	}
	got = mapNumbers(t, string(readme), "## The numbering is the documentation", "```\n\nSteps sharing")
	assertSameNumbers(t, "the step map in README.md", got, want)
}

// assertSameNumbers reports what is missing from or extra in a map.
func assertSameNumbers(t *testing.T, where string, got, want []int) {
	t.Helper()

	inWant := map[int]bool{}
	for _, n := range want {
		inWant[n] = true
	}
	inGot := map[int]bool{}
	for _, n := range got {
		inGot[n] = true
	}
	for _, n := range want {
		if !inGot[n] {
			t.Errorf("%s does not list step %d, which steps() registers", where, n)
		}
	}
	for _, n := range got {
		if !inWant[n] {
			t.Errorf("%s lists step %d, which steps() does not register", where, n)
		}
	}
}

// TestREADMEStepMapIsInOrder checks the README's map is sorted.
//
// The document states two lines below it that "a higher number needs
// the lower ones", so a map listing 35 after 40 contradicts the rule it
// is illustrating.
func TestREADMEStepMapIsInOrder(t *testing.T) {
	t.Parallel()

	readme, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("reading the README: %v", err)
	}
	i := strings.Index(string(readme), "## The numbering is the documentation")
	if i < 0 {
		t.Fatal("could not find the step map heading")
	}
	rest := string(readme)[i:]
	if j := strings.Index(rest, "```\n\nSteps sharing"); j >= 0 {
		rest = rest[:j]
	}

	var seq []int
	for _, line := range strings.Split(rest, "\n") {
		m := stepLineRE.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		var n int
		if _, err := fmt.Sscanf(m[1], "%d", &n); err == nil {
			seq = append(seq, n)
		}
	}
	for i := 1; i < len(seq); i++ {
		if seq[i] < seq[i-1] {
			t.Errorf("the README step map lists %d after %d; it must read in ascending order, since the text below it says a higher number needs the lower ones", seq[i], seq[i-1])
		}
	}
}
