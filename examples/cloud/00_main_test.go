package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/sitehostnz/gosh/pkg/api/server"
	"github.com/sitehostnz/gosh/pkg/shtypes"
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
// program, which has happened twice in the server journey — and the
// second time three of the undocumented steps wrote to real
// infrastructure.
//
// Numbers are compared rather than descriptions: the wording is meant
// to differ between a godoc table and a README, but a number in one and
// not the other is always a mistake.
func TestStepMapsMatchTheSteps(t *testing.T) {
	t.Parallel()

	want := registeredNumbers()

	doc, err := os.ReadFile("00_main.go")
	if err != nil {
		t.Fatalf("reading the package doc: %v", err)
	}
	got := mapNumbers(t, string(doc), "// # Reading the journey", "// Run with no arguments")
	assertSameNumbers(t, "the package doc comment in 00_main.go", got, want)

	readme, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("reading the README: %v", err)
	}
	got = mapNumbers(t, string(readme), "## The steps", "```\n\nRun with no arguments")
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

// TestEveryWritingStepClaimsAnOutOfBandCheck is the one that guards this
// journey's actual promise.
//
// Both the package doc and the README state that every writing step is
// verified outside the control plane, and then list them. That claim is
// the reason the journey provisions anything: a step that writes and
// then only asks the API whether the write happened is the failure this
// journey exists to avoid. A new mutating step added without a line in
// those tables would make the claim false in the two places a reader
// meets it first.
//
// This checks the tables mention every mutating step by name. It cannot
// check the step really verifies out of band — but it makes the
// omission impossible to make silently.
func TestEveryWritingStepClaimsAnOutOfBandCheck(t *testing.T) {
	t.Parallel()

	doc, err := os.ReadFile("00_main.go")
	if err != nil {
		t.Fatalf("reading the package doc: %v", err)
	}
	readme, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("reading the README: %v", err)
	}

	docTable := section(t, string(doc), "// # Every write is verified outside the control plane", "// # The label is not the name")
	readmeTable := section(t, string(readme), "## Every write is checked outside the control plane", "### Public routing works, but a malformed stack can break it")

	for _, s := range steps() {
		if !s.mutates {
			continue
		}
		if !strings.Contains(docTable, s.name) {
			t.Errorf("the package doc's out-of-band table does not mention the writing step %q", s.name)
		}
		if !strings.Contains(readmeTable, s.name) {
			t.Errorf("the README's out-of-band table does not mention the writing step %q", s.name)
		}
	}
}

// section returns the text between two markers, failing if either is
// absent — an absent marker means the document was restructured and
// the check silently stopped looking at anything.
func section(t *testing.T, body, from, to string) string {
	t.Helper()

	i := strings.Index(body, from)
	if i < 0 {
		t.Fatalf("could not find %q", from)
	}
	rest := body[i+len(from):]
	j := strings.Index(rest, to)
	if j < 0 {
		t.Fatalf("could not find %q after %q", to, from)
	}
	return rest[:j]
}

// TestComposeSetsVirtualHost pins the routing contract.
//
// VIRTUAL_HOST is what the host's reverse proxy routes on, so a stack
// without it deploys successfully and is unreachable. Step 40's HTTP
// check would then fail with nothing to indicate why, so the compose
// file is asserted directly.
func TestComposeSetsVirtualHost(t *testing.T) {
	t.Parallel()

	st := &state{stackName: "cc0123456789abcd", stackHost: "cc0123456789abcd.203.0.113.7.sth.nz"}
	got := composeFor(st)

	for _, want := range []string{
		"VIRTUAL_HOST=" + st.stackHost,
		"container_name: " + st.stackName,
		"    " + st.stackName + ":",
		"nz.sitehost.container.type=www",
		"expose:",
		// The proxy lives on this network. A container that does not
		// join it reports Up and answers nothing but the proxy's 503,
		// which is what happened before this line existed.
		"name: infra_default",
		// The image serves from here, and step 50 writes its probe
		// page into it.
		"/container/application:rw",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("the compose file does not contain %q:\n%s", want, got)
		}
	}

	// "ports" would publish to the host, which collides between stacks
	// and is unnecessary because the proxy uses the Docker network.
	if strings.Contains(got, "ports:") {
		t.Errorf("the compose file publishes ports to the host; it should only expose:\n%s", got)
	}
}

// TestShellQuoteEscapes checks a generated secret cannot break out of
// the shell word it is embedded in.
//
// Passwords here are hex, so this cannot bite today. It is pinned
// because the generator is one edit away from including punctuation,
// and a quote in a password would turn the database probe into a
// different command — which is both a broken check and a shell
// injection into a host this journey has a login on.
func TestShellQuoteEscapes(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct{ in, want string }{
		{"plain", "'plain'"},
		{"it's", `'it'"'"'s'`},
		{"a b", "'a b'"},
		{`; rm -rf /`, `'; rm -rf /'`},
	} {
		if got := shellQuote(tc.in); got != tc.want {
			t.Errorf("shellQuote(%q) = %s, want %s", tc.in, got, tc.want)
		}
	}
}

// TestSmallerOrdersByCoresThenPrice checks product selection is
// deterministic.
//
// AKLNCT offers several container codes with the same core count, so
// cores alone would leave the choice to whichever the API listed first
// — and a journey that provisions a different product on different days
// is one whose failures are hard to reproduce.
func TestSmallerOrdersByCoresThenPrice(t *testing.T) {
	t.Parallel()

	mk := func(code string, cores int, price string) server.Product {
		var p server.Product
		p.Code = code
		p.Attributes.Cores = cores
		p.Price = shtypes.MaybeString(price)
		return p
	}

	one := mk("CLDCON1", 1, "35")
	oneDear := mk("CLDCON1-P", 1, "55")
	two := mk("CLDCON2", 2, "80")

	if !smaller(one, two) {
		t.Error("a 1-core product should sort smaller than a 2-core one")
	}
	if smaller(two, one) {
		t.Error("a 2-core product should not sort smaller than a 1-core one")
	}
	if !smaller(one, oneDear) {
		t.Error("with equal cores the cheaper product should sort smaller")
	}
	if smaller(oneDear, one) {
		t.Error("with equal cores the dearer product should not sort smaller")
	}
	// Irreflexivity: a product is not smaller than itself, or the
	// selection loop could keep replacing its own choice.
	if smaller(one, one) {
		t.Error("a product should not sort smaller than itself")
	}
}

// TestGatewayErrorSeparatesProxyFromContainer pins the distinction the
// serving check depends on.
//
// The host's reverse proxy answers on port 80 whether or not a
// container is behind it, returning its own 503 when nothing is
// routable. So "something answered" is satisfied by a stack that
// deployed and never started — and the first version of the stack check
// accepted exactly that, reporting a pass on a 503 while its comment
// claimed to catch the case.
func TestGatewayErrorSeparatesProxyFromContainer(t *testing.T) {
	t.Parallel()

	for _, code := range []int{
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout,
	} {
		if !gatewayError(code) {
			t.Errorf("HTTP %d should be treated as the proxy failing to reach a container", code)
		}
	}

	// These come from the container, which is what the check
	// establishes. 403 and 404 matter specifically: a PHP image with an
	// empty docroot answers one of them, and that is a container
	// answering.
	for _, code := range []int{
		http.StatusOK,
		http.StatusMovedPermanently,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusInternalServerError,
	} {
		if gatewayError(code) {
			t.Errorf("HTTP %d comes from the container and should be accepted as it answering", code)
		}
	}
}

// TestTransientRefusalsAreRecognised pins the teardown's retry
// predicate.
//
// Three of the five teardown parts failed in one run on messages that
// meant "not yet": a job still queued on a stack, an SSH user with
// sessions still open, and a lock. All three would have succeeded
// moments later, and a teardown that gives up on the first one leaves
// real infrastructure behind.
//
// The negative cases matter as much: retrying a genuine "not found" or
// a permissions error would turn a fast failure into a four-minute one
// and then report the same thing.
func TestTransientRefusalsAreRecognised(t *testing.T) {
	t.Parallel()

	transient := []string{
		"Error: Unable to delete stack, there is a job already running on this stack.",
		"[gosh1] Sorry, this SSH user cannot be deleted right now, please ensure all sessions are logged out, then try again.",
		"Error: Unable to create database, the server is currently locked.",
	}
	for _, msg := range transient {
		if !isTransientRefusal(errors.New(msg)) {
			t.Errorf("should be retried, was not: %s", msg)
		}
	}

	permanent := []string{
		"Error: Unable to remove database user, the database user couldn't be found or you don't have permission to delete it.",
		"Error: Not Found",
		"Error: The mysql_host parameter is missing.",
	}
	for _, msg := range permanent {
		if isTransientRefusal(errors.New(msg)) {
			t.Errorf("should not be retried, was: %s", msg)
		}
	}

	if isTransientRefusal(nil) {
		t.Error("a nil error is not a refusal")
	}
}
