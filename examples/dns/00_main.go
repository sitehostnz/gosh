// Program dns walks the SiteHost DNS endpoints as a numbered journey,
// recording every exchange with the API as it goes.
//
// # Why this exists
//
// The dns and dns/template packages between them wrap thirty methods
// and had no example. Nothing exercised the template package at all,
// which is why its behaviour is undocumented: an endpoint nobody calls
// has no observed quirks, only assumed ones.
//
// # Recording is the point, not a side effect
//
// Set SH_RECORD_DIR and every call is written out as JSON, including
// the ones that fail.
//
// A hand-written mock encodes what we believe the API accepts, so a
// test built on one can only confirm the belief that produced it.
// Several bugs in this SDK survived a green suite that way — a
// response field declared a bool that the API sends as a map, a
// parameter added under a name the encoder then dropped. Every one of
// those tests asserted what we sent, and passed while the real call
// failed.
//
// Rejections are the valuable half of a recording corpus, and the
// cheapest to collect: address something that cannot exist and nothing
// is created, so nothing has to be cleaned up.
//
// # Reading the journey
//
//	10  discover   read-only; safe anywhere
//	20  zone       create a DNS zone
//	30  records    add, read back, change and remove a record
//	40  template   relink the zone to another template, then restore
//	80  probe      deliberate rejections, read-only, no opt-in
//	90  delete     always last
//
// A higher number needs the lower ones. Run with no arguments to print
// the map and exit.
//
// # What this journey cannot check, and why
//
// Every assertion here reads back through the same API that made the
// change. That proves a record changed, not that anything happened,
// and elsewhere in this repository the answer is to look somewhere
// else — a socket for a firewall rule, a file in the guest for a
// snapshot restore.
//
// The obvious equivalent is a DNS query against the authoritative
// nameservers. It does not work, and that is worth recording rather
// than leaving as an absence. Asked directly, ns1.sitehost.co.nz
// answers for a zone SiteHost hosts:
//
//	dig +short @ns1.sitehost.co.nz sitehost.co.nz SOA
//	ns1.sitehost.co.nz. soa.sitehost.co.nz. 2026081702 ...
//
// and returns nothing at all for a zone this journey has just created
// and confirmed through the API. So creating a zone does not by itself
// make it resolvable: something further along — delegation, or a
// publish step this API does not expose — has to happen first.
//
// Which means an out-of-band check is not available here, and claiming
// one would be worse than admitting it. If that changes, a resolver
// query against a created zone is the check to add.
//
// # Creating a zone is not registering a domain
//
// They are separate operations that both get called "adding a domain".
// This journey does the first, which is free. It defaults to a name
// under .invalid — reserved by RFC 2606 so it can never be registered
// — so a mistake cannot collide with anything real.
//
// Required env: SH_API_KEY, SH_CLIENT_ID.
// Required to create anything: SH_EXAMPLE_ALLOW_PROVISION=1.
// Optional: SH_ZONE, SH_BASE_URL, SH_RECORD_DIR, SH_DELETE_ZONE.
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
			describe: "list the zones and reverse-DNS addresses",
		},
		{
			order: 20, name: "zone", inTour: true, mutates: true, run: stepZone,
			needs:    "nothing; refuses a zone it did not create",
			describe: "create the DNS zone the later steps act on",
		},
		{
			order: 30, name: "records", inTour: true, mutates: true, run: stepRecords,
			needs:    "step 20",
			describe: "add, read back, change and remove a record",
		},
		{
			order: 40, name: "template", inTour: true, mutates: true, run: stepTemplate,
			needs:    "step 20",
			describe: "relink the zone to another template, then restore",
		},
		{
			order: 80, name: "probe", inTour: true, run: stepProbe,
			needs:    "nothing; read-only",
			describe: "provoke rejections deliberately and record them",
		},
		{
			order: 90, name: "delete", mutates: true, run: stepDelete,
			needs:    "nothing; always safe to run last",
			describe: "delete the zone this run created",
		},
	}
}

func main() {
	log.SetFlags(log.Ltime)
	if err := run(os.Args[1:]); err != nil {
		log.Fatalf("dns: %v", err)
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
	fmt.Println("SiteHost DNS journey")
	fmt.Println()
	fmt.Printf("  %-5s %-10s %-7s %-46s %s\n", "STEP", "NAME", "WRITES", "WHAT IT DOES", "NEEDS")
	fmt.Printf("  %-5s %-10s %-7s %-46s %s\n", "-----", "----------", "-------",
		strings.Repeat("-", 46), strings.Repeat("-", 24))
	for _, s := range all {
		writes := "no"
		if s.mutates {
			writes = "YES"
		}
		fmt.Printf("  %-5d %-10s %-7s %-46s %s\n", s.order, s.name, writes, s.describe, s.needs)
	}
	fmt.Println()
	fmt.Println("Record every call, including the rejections:")
	fmt.Println("  SH_RECORD_DIR=./recordings go run ./examples/dns probe")
	fmt.Println()
	fmt.Println("Steps marked WRITES need SH_EXAMPLE_ALLOW_PROVISION=1.")
}

// runJourney runs the tour, always attempting cleanup.
func runJourney(ctx context.Context, c clients, st *state, all []journeyStep) error {
	if !allowed() && anyMutates(all) {
		return fmt.Errorf("the journey creates a DNS zone; set SH_EXAMPLE_ALLOW_PROVISION=1 (run with no arguments for the map)")
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

// runTour runs every step but the cleanup.
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
			return fmt.Errorf("step %q changes real DNS; set SH_EXAMPLE_ALLOW_PROVISION=1", name)
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

// allowed reports whether the caller opted in to changing anything.
func allowed() bool { return os.Getenv("SH_EXAMPLE_ALLOW_PROVISION") == "1" }
