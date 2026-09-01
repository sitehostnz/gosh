// Program cloud walks the Cloud Container lifecycle, as a numbered
// journey, recording every exchange with the API as it goes.
//
// # Why this exists
//
// The cloud packages had no example and no test coverage at all —
// cloud/db, cloud/db/user, cloud/db/grant, cloud/ssh/user and
// cloud/stack/image all sat at 0.0%. Nothing had ever exercised them,
// which is why their behaviour is undocumented: an endpoint nobody
// calls has no observed quirks, only assumed ones.
//
// # Recording is the point, not a side effect
//
// Set SH_RECORD_DIR and every call is written out as JSON, including
// the ones that fail. That matters more than it sounds:
//
// A hand-written mock encodes what we believe the API accepts, so a
// test built on one can only confirm the belief that produced it. Two
// bugs in this SDK survived a green suite that way — a provision that
// sent an array where a scalar was required, and a decode that declared
// a map as a bool. Both tests asserted what we sent, and passed while
// every real call failed.
//
// Recordings of REJECTIONS are the valuable half. A corpus of successes
// only says what came back for requests that already worked; neither bug
// above would appear in one. Rejections are also the cheapest thing to
// collect, since nothing is provisioned and nothing has a side effect.
//
// # Reading the journey
//
//	10  discover           read-only; safe anywhere
//	20  provision          creates a Cloud Container
//	30  stack              deploy a container stack
//	40  database           add a database and a user
//	50  sshuser            add an SSH user
//	80  probe              deliberate rejections, read-only, no opt-in
//	90  delete             always last
//
// Steps sharing a number are interchangeable; a higher number needs the
// lower ones. Run with no arguments to print the map and exit.
//
// Required env: SH_API_KEY, SH_CLIENT_ID.
// Required to create anything: SH_EXAMPLE_ALLOW_PROVISION=1.
// Optional: SH_LOCATION, SH_PRODUCT, SH_SERVER, SH_RECORD_DIR.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
)

// stepFn is one step of the journey.
type stepFn func(context.Context, clients, *state) error

// journeyStep pairs a step with its ordering and what it needs.
type journeyStep struct {
	order    int
	name     string
	needs    string
	mutates  bool
	inTour   bool
	run      stepFn
	describe string
}

func steps() []journeyStep {
	return []journeyStep{
		{
			order: 10, name: "discover", inTour: true, run: stepDiscover,
			needs:    "nothing",
			describe: "find a location and product that offer containers",
		},
		{
			order: 40, name: "read", inTour: true, run: stepRead,
			needs:    "an existing cloud server; SH_SERVER, or the first on the account",
			describe: "walk every read path and check the shapes agree",
		},
		{
			order: 80, name: "probe", inTour: true, run: stepProbe,
			needs:    "nothing; read-only",
			describe: "provoke rejections deliberately and record them",
		},
	}
}

func main() {
	log.SetFlags(log.Ltime)
	if err := run(os.Args[1:]); err != nil {
		log.Fatalf("cloud: %v", err)
	}
}

func run(args []string) error {
	all := steps()
	if len(args) == 0 {
		printMap(all)
		return nil
	}

	c, err := newClients()
	if err != nil {
		return err
	}
	st := &state{cfg: newConfig()}
	ctx := context.Background()

	if args[0] == "journey" {
		return runJourney(ctx, c, st, all)
	}
	return runOne(ctx, c, st, all, args[0])
}

// printMap prints the journey without touching the API.
func printMap(all []journeyStep) {
	fmt.Println("SiteHost Cloud Container journey")
	fmt.Println()
	fmt.Printf("  %-5s %-10s %-7s %-46s %s\n", "STEP", "NAME", "WRITES", "WHAT IT DOES", "NEEDS")
	fmt.Printf("  %-5s %-10s %-7s %-46s %s\n", "-----", "----------", "-------", strings.Repeat("-", 46), strings.Repeat("-", 24))
	for _, s := range all {
		writes := "no"
		if s.mutates {
			writes = "YES"
		}
		fmt.Printf("  %-5d %-10s %-7s %-46s %s\n", s.order, s.name, writes, s.describe, s.needs)
	}
	fmt.Println()
	fmt.Println("Record every call, including the rejections:")
	fmt.Println("  SH_RECORD_DIR=$(mktemp -d) go run ./examples/cloud probe")
	fmt.Println()
	fmt.Println("Steps marked WRITES need SH_EXAMPLE_ALLOW_PROVISION=1.")
}

// runJourney runs the tour, always attempting cleanup.
func runJourney(ctx context.Context, c clients, st *state, all []journeyStep) error {
	if !allowed() && anyMutates(all) {
		return fmt.Errorf("the journey creates real resources; set SH_EXAMPLE_ALLOW_PROVISION=1 (run with no arguments for the map)")
	}
	runErr := runTour(ctx, c, st, all)
	if err := runCleanup(ctx, c, st, all); err != nil {
		if runErr != nil {
			return fmt.Errorf("%w (cleanup also failed: %v)", runErr, err)
		}
		return err
	}
	if runErr != nil {
		return runErr
	}
	log.Printf("✓ journey complete")
	return nil
}

// runTour runs every step but the cleanup, stopping at the first
// failure so that cleanup still sees the state it left behind.
func runTour(ctx context.Context, c clients, st *state, all []journeyStep) error {
	for _, s := range all {
		if !s.inTour || s.name == "delete" {
			continue
		}
		stepf("%d %s — %s", s.order, s.name, s.describe)
		if err := s.run(ctx, c, st); err != nil {
			return fmt.Errorf("step %d %s: %w", s.order, s.name, err)
		}
	}
	return nil
}

// runCleanup runs the delete step, whether or not the tour succeeded.
func runCleanup(ctx context.Context, c clients, st *state, all []journeyStep) error {
	for _, s := range all {
		if s.name != "delete" {
			continue
		}
		stepf("%d %s — %s", s.order, s.name, s.describe)
		if err := s.run(ctx, c, st); err != nil {
			return err
		}
	}
	return nil
}

// runOne runs a single named step.
func runOne(ctx context.Context, c clients, st *state, all []journeyStep, name string) error {
	for _, s := range all {
		if s.name != name {
			continue
		}
		if s.mutates && !allowed() {
			return fmt.Errorf("step %q creates real resources; set SH_EXAMPLE_ALLOW_PROVISION=1", name)
		}
		stepf("%d %s — %s", s.order, s.name, s.describe)
		return s.run(ctx, c, st)
	}
	names := make([]string, 0, len(all))
	for _, s := range all {
		names = append(names, s.name)
	}
	return fmt.Errorf("unknown step %q; known: %s (or \"journey\")", name, strings.Join(names, ", "))
}

// anyMutates reports whether the tour contains a writing step.
func anyMutates(all []journeyStep) bool {
	for _, s := range all {
		if s.inTour && s.mutates {
			return true
		}
	}
	return false
}

// allowed reports whether the caller opted in to creating resources.
func allowed() bool { return os.Getenv("SH_EXAMPLE_ALLOW_PROVISION") == "1" }
