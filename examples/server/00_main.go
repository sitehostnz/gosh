// Program server walks the whole lifecycle of a SiteHost virtual
// server through the API, as a numbered journey.
//
// # Reading the journey
//
// The files in this directory are numbered by the order they must run
// in:
//
//	10  register an SSH key      must precede provisioning
//	20  discover and preflight   before provisioning, any order
//	30  provision the pair       needs 10 and 20
//	35  upgrades (opt-in)        needs 30; not run by "journey"
//	40  firewall / netconfig     after provisioning, any order
//	50  prestage the guests      needs the key from 10; MUST precede 60
//	60  swap the addresses       needs a provisioned pair
//	70  reboot via the API       needs 50 and 60
//	80  verify over SSH          needs 70
//	90  delete                   always last
//
// The 50/60/70 ordering is the one that is not obvious. A guest cannot
// be reached once its address has moved, so its configuration has to be
// staged beforehand — and the reboot that applies it goes through the
// API, which needs no access to the guest at all.
//
// Steps sharing a number are interchangeable: run them in any order,
// or skip them. A higher number depends on the lower ones having
// happened. That is the whole point of the numbering — the ordering
// constraints of the API are not otherwise written down anywhere, and
// they are the part that catches people out.
//
// Run with no arguments to print the map and exit. Nothing is created
// and nothing is charged:
//
//	go run ./examples/server
//
// Run one step, against servers you already have:
//
//	SH_SERVER_A=... SH_SERVER_B=... go run ./examples/server netconfig
//
// Run the whole thing, which provisions two real servers and deletes
// them again:
//
//	SH_EXAMPLE_ALLOW_PROVISION=1 go run ./examples/server journey
//
// # Why this is one program rather than many
//
// Every step needs the same plumbing: per-second rate limiting, job
// polling, the fact that a server's name is not its label, and a
// teardown that runs even when an assertion fails. Written as separate
// programs that plumbing gets copied six times and drifts. Here it
// lives once, in 00_shared.go.
//
// # Assertions, not a demo
//
// Every step checks something that can actually fail and exits non-zero
// when it does. A step that cannot fail is worse than no step: it reads
// as coverage.
//
// Required env: SH_API_KEY, SH_CLIENT_ID.
// Required to mutate anything: SH_EXAMPLE_ALLOW_PROVISION=1.
// Optional env: SH_LOCATION, SH_PRODUCT, SH_IMAGE, SH_DISTRO,
// SH_LABEL_A, SH_LABEL_B, SH_SERVER_A, SH_SERVER_B, SH_BASE_URL,
// SH_SSH_KEY_FILE, SH_DISK_LABEL, SH_UPGRADE_PLAN, SH_DELETE_SERVERS.
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

// journeyStep pairs a step with its ordering number and a one-line
// description of what it needs.
type journeyStep struct {
	order    int
	name     string
	needs    string
	mutates  bool
	run      stepFn
	inTour   bool // included when the full journey runs
	describe string
}

// steps is the journey. Order in this slice is the run order for
// "journey"; the order field is what the map prints.
func steps() []journeyStep {
	out := make([]journeyStep, 0, 12)
	out = append(out, setupSteps()...)
	out = append(out, upgradeSteps()...)
	out = append(out, cutoverSteps()...)
	return out
}

// setupSteps are everything up to and including provisioning.
func setupSteps() []journeyStep {
	return []journeyStep{
		{
			order: 10, name: "sshkey", mutates: true, inTour: true, run: stepSSHKey,
			needs:    "nothing",
			describe: "generate a keypair and register the public half",
		},
		{
			order: 20, name: "discover", inTour: true, run: stepDiscover,
			needs:    "nothing",
			describe: "list locations and images, resolve an image code",
		},
		{
			order: 20, name: "preflight", inTour: true, run: stepPreflight,
			needs:    "a location and product",
			describe: "check the product has capacity at the location",
		},
		{
			order: 25, name: "inventory", inTour: true, run: stepInventory,
			needs:    "nothing; read-only",
			describe: "read every account-level listing the SDK exposes",
		},
		{
			order: 30, name: "provision", mutates: true, inTour: true, run: stepProvision,
			needs:    "steps 10 and 20",
			describe: "create two servers and wait for both jobs",
		},
	}
}

// upgradeSteps change what a server costs or how big it is, so they are
// opt-in rather than part of the journey.
func upgradeSteps() []journeyStep {
	return []journeyStep{
		{
			order: 35, name: "upgrade-disk", mutates: true, run: stepUpgradeDisk,
			needs:    "a provisioned server; opt-in, not run by \"journey\"",
			describe: "stage a disk resize, then commit it (online on HPVS)",
		},
		{
			order: 35, name: "upgrade-plan", mutates: true, run: stepUpgradePlan,
			needs:    "SH_UPGRADE_PLAN; opt-in, not run by \"journey\"",
			describe: "move the server to another product code",
		},
	}
}

// cutoverSteps inspect, swap, reboot, verify and clean up.
func cutoverSteps() []journeyStep {
	return []journeyStep{
		{
			order: 40, name: "firewall", inTour: true, run: stepFirewall,
			needs:    "a provisioned server",
			describe: "report groups and whether inbound SSH is open",
		},
		{
			order: 40, name: "netconfig", inTour: true, run: stepNetConfig,
			needs:    "a provisioned server",
			describe: "fetch the guest network configuration files",
		},
		{
			order: 45, name: "secgroup", mutates: true, inTour: true, run: stepSecGroup,
			needs:    "a provisioned server; high-performance only",
			describe: "create a security group, attach, change rules, remove",
		},
		{
			order: 47, name: "snapshot", mutates: true, inTour: true, run: stepSnapshot,
			needs:    "a server this run provisioned",
			describe: "take a snapshot, set its lifetime, restore from it",
		},
		{
			order: 48, name: "label", mutates: true, inTour: true, run: stepLabel,
			needs:    "a server this run provisioned",
			describe: "change a server's label and put it back",
		},
		{
			order: 50, name: "prestage", mutates: true, inTour: true, run: stepPrestage,
			needs:    "step 10's key and a provisioned pair",
			describe: "stage each guest's post-swap config, unapplied",
		},
		{
			order: 60, name: "swap", mutates: true, inTour: true, run: stepSwap,
			needs:    "a provisioned pair",
			describe: "swap the two primary addresses",
		},
		{
			order: 70, name: "reboot", mutates: true, inTour: true, run: stepReboot,
			needs:    "steps 50 and 60",
			describe: "reboot both via the API so the staged config applies",
		},
		{
			order: 80, name: "verify", inTour: true, run: stepVerify,
			needs:    "a completed reboot",
			describe: "log in on the swapped addresses and confirm identity",
		},
		{
			order: 90, name: "delete", mutates: true, inTour: true, run: stepDelete,
			needs:    "nothing; always safe to run last",
			describe: "delete everything this run created",
		},
	}
}

func main() {
	log.SetFlags(log.Ltime)
	if err := run(os.Args[1:]); err != nil {
		log.Fatalf("server: %v", err)
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

// printMap prints the journey without touching the API. This is the
// default output because it is the thing a reader — or an agent — needs
// first: what has to happen, and in what order.
func printMap(all []journeyStep) {
	fmt.Println("SiteHost server journey")
	fmt.Println()
	fmt.Println("Steps sharing a number run in any order. A higher number needs the lower ones.")
	fmt.Println()
	fmt.Printf("  %-4s %-10s %-7s %-52s %s\n", "STEP", "NAME", "WRITES", "WHAT IT DOES", "NEEDS")
	fmt.Printf("  %-4s %-10s %-7s %-52s %s\n", "----", "----------", "-------", strings.Repeat("-", 52), strings.Repeat("-", 38))
	for _, s := range all {
		writes := "no"
		if s.mutates {
			writes = "YES"
		}
		fmt.Printf("  %-4d %-10s %-7s %-52s %s\n", s.order, s.name, writes, s.describe, s.needs)
	}
	fmt.Println()
	fmt.Println("Run the whole journey (provisions two real servers, then deletes them):")
	fmt.Println("  SH_EXAMPLE_ALLOW_PROVISION=1 go run ./examples/server journey")
	fmt.Println()
	fmt.Println("Run one step against servers you already have:")
	fmt.Println("  SH_SERVER_A=... SH_SERVER_B=... go run ./examples/server netconfig")
	fmt.Println()
	fmt.Println("Read-only steps need no opt-in. Steps marked WRITES require")
	fmt.Println("SH_EXAMPLE_ALLOW_PROVISION=1 and create or change real infrastructure.")
	fmt.Println()
	fmt.Println("Steps whose NEEDS says opt-in are not run by \"journey\" — they change")
	fmt.Println("what a server costs or how big it is. Run them by name.")
	fmt.Println()
	fmt.Println("Two things worth knowing before you start:")
	fmt.Println("  - product codes ARE discoverable, via server.ListProducts —")
	fmt.Println("    an undocumented endpoint, which is why the KB page listing")
	fmt.Println("    them is maintained by hand. Location is required.")
	fmt.Println("  - a server's name is not its label: the platform truncates the")
	fmt.Println("    label and appends a digit on collision. Always use the name")
	fmt.Println("    returned by the provision call.")
}

// runJourney runs every step in order, guaranteeing the delete step
// runs even when something before it fails.
func runJourney(ctx context.Context, c clients, st *state, all []journeyStep) error {
	if !allowed() {
		return fmt.Errorf("the full journey provisions real servers; set SH_EXAMPLE_ALLOW_PROVISION=1 to proceed (run with no arguments to see the map)")
	}

	runErr := runTour(ctx, c, st, all)

	// Cleanup is always attempted, so a failure part-way cannot leak a
	// server. Both failures are surfaced if cleanup breaks too.
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

// runTour runs every step except the delete.
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

// runCleanup runs the delete step, wherever it sits in the list.
func runCleanup(ctx context.Context, c clients, st *state, all []journeyStep) error {
	for _, s := range all {
		if s.name != "delete" {
			continue
		}
		stepf("%d %s — %s", s.order, s.name, s.describe)
		if err := s.run(ctx, c, st); err != nil {
			return fmt.Errorf("step %d delete: %w", s.order, err)
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
			return fmt.Errorf("step %q changes real infrastructure; set SH_EXAMPLE_ALLOW_PROVISION=1 to proceed", name)
		}
		stepf("%d %s — %s", s.order, s.name, s.describe)
		return s.run(ctx, c, st)
	}
	names := make([]string, 0, len(all))
	for _, s := range all {
		names = append(names, s.name)
	}
	return fmt.Errorf("unknown step %q; known steps: %s (or \"journey\")", name, strings.Join(names, ", "))
}

// allowed reports whether the caller has opted in to changing real
// infrastructure.
func allowed() bool {
	return os.Getenv("SH_EXAMPLE_ALLOW_PROVISION") == "1"
}
