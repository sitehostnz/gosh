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
// test built on one can only confirm the belief that produced it.
// Bugs in this SDK have survived a green suite that way. A response
// field declared a bool where the API sends a map meant every disk
// upgrade failed to decode, and the fixture asserting the wrong shape
// passed throughout. Parameters have gone out under names the encoder
// then dropped, so the call succeeded and the value never arrived —
// four of those, each with a passing test.
//
// Recordings of REJECTIONS are the valuable half. A corpus of successes
// only says what came back for requests that already worked; neither bug
// above would appear in one. Rejections are also the cheapest thing to
// collect, since nothing is provisioned and nothing has a side effect.
//
// # Reading the journey
//
//	10  discover           read-only; safe anywhere
//	20  provision  WRITES  create a container server, prove it answers
//	30  stack      WRITES  deploy a website container, fetch it
//	40  sshuser    WRITES  add a container SSH user, log into it
//	50  database   WRITES  create a database, user and grants
//	60  read               walk every read path and check the shapes
//	80  probe              deliberate rejections, read-only, no opt-in
//	90  delete     WRITES  tear down what this process created
//
// Run with no arguments to print the map and exit.
//
// # Every write is verified outside the control plane
//
// This is the point of the journey, and the reason it provisions
// anything at all. Asking the API whether the API did what it said
// proves that a record changed, not that anything happened. Each
// writing step therefore has a check that does not go through the
// control plane:
//
//	20  provision  a TCP handshake to tcp/22 completes
//	30  stack      the compose file round-trips; serving is proven in 40
//	40  sshuser    an SSH session opens as that account, writes a marker into
//	               the container, and fetches it back over the Docker network
//	50  database   the mysql client, run inside the host as the created user,
//	               names the database, writes a row and reads it back
//	90  delete     tcp/22 stops answering
//
// A completed job, an "Up" state and a returned record are all reported
// too, because a disagreement between them and reality is worth
// catching. None of them is the assertion.
//
// # The label is not the name
//
// A container server's name is derived from the provision label, not
// equal to it: the platform prefixes "ch-" and truncates, so the label
// "gosh-cloud-journey" produced the name "ch-gosh-clou". Collisions get
// a digit appended. Everything after step 20 uses the returned name.
//
// # Ports belong to the container's type, not to the compose file
//
// As of writing, against API 1.5, it is not possible to publish
// ports on a www container — which is what this journey deploys. 80 and
// 443 are open and that is all.
// A compose file carrying a "ports" mapping is accepted, keeps the
// mapping when read back, and comes up with the container running,
// while the port stays shut; a restart does not change it. Service
// containers are the ones that publish, subject to a reserved range
// and a specific block list.
//
// This is the single most misleading behaviour in these endpoints,
// because every observable signal short of the socket itself says it
// worked. See step 30 for the detail, and
// https://kb.sitehost.nz/cloud-containers/containers/ports for the
// rules.
//
// # What this journey does not cover
//
// Stack copy and backup, volumes, SSL, and postgres database stacks are
// untouched. The database step deploys a MariaDB stack because a fresh
// container server has none, so which database engines behave
// identically here is unverified — only the one this journey deploys is
// exercised.
//
// Required env: SH_API_KEY, SH_CLIENT_ID.
// Optional: SH_LOCATION, SH_PRODUCT, SH_IMAGE, SH_LABEL, SH_SERVER,
// SH_BASE_URL, SH_RECORD_DIR.
//
// Steps that write require SH_EXAMPLE_ALLOW_PROVISION=1. Only servers
// this process provisioned are ever deleted; a server named by
// SH_SERVER is never a delete target.
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
			describe: "find a location and resolve the smallest container product",
		},
		{
			order: 20, name: "provision", inTour: true, mutates: true, run: stepProvision,
			needs:    "step 10",
			describe: "create a container server; prove tcp/22 answers",
		},
		{
			order: 30, name: "stack", inTour: true, mutates: true, run: stepStack,
			needs:    "a container server; state or SH_SERVER",
			describe: "deploy a website container; fetch it over HTTP",
		},
		{
			order: 40, name: "sshuser", inTour: true, mutates: true, run: stepSSHUser,
			needs:    "step 30, same process",
			describe: "add a container SSH user; log into it over SSH",
		},
		{
			order: 50, name: "database", inTour: true, mutates: true, run: stepDatabase,
			needs:    "steps 30 and 40, same process",
			describe: "create a database, user and grants; connect as that user",
		},
		{
			order: 60, name: "read", inTour: true, run: stepRead,
			needs:    "a cloud server; state, SH_SERVER, or the first on the account",
			describe: "walk every read path and check the shapes agree",
		},
		{
			order: 80, name: "probe", inTour: true, run: stepProbe,
			needs:    "nothing; read-only",
			describe: "provoke rejections deliberately and record them",
		},
		{
			order: 90, name: "delete", inTour: true, mutates: true, run: stepDelete,
			needs:    "step 20, same process",
			describe: "tear down what this process created; prove it stops answering",
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
	// Several names run in one process, in the order given. Steps hand
	// state to each other — the stack name, the address, the SSH
	// account — so "stack sshuser database" is the only way to exercise
	// those three against a server that already exists. Running them as
	// separate processes cannot work, and the error when you try says
	// only that something is missing from state.
	for _, name := range args {
		if err := runOne(ctx, c, st, all, name); err != nil {
			return err
		}
	}
	return nil
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
	fmt.Println()
	fmt.Println("Several steps run in one process, sharing state:")
	fmt.Println("  SH_SERVER=<name> go run ./examples/cloud stack sshuser database")
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
