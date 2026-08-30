package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

const (
	// verifyTimeout bounds how long to wait for a guest to come back on
	// its new address.
	verifyTimeout = 8 * time.Minute

	// verifySettle is how long to wait before polling at all, so the
	// reboot has actually begun. Without it the first connection can
	// land on the guest as it was before the reboot.
	verifySettle = 20 * time.Second
)

// stepVerify proves the cutover worked, by logging in on the swapped
// addresses and checking which machine answers.
//
// # Reachability is not the assertion
//
// Something answering on an address proves only that *a* machine has
// it — and right after a swap that is quite likely the wrong one. The
// guest that used to own the address still has it configured until it
// reboots, so an early connection reaches the outgoing server and
// everything looks fine.
//
// So the check is on identity, not reachability: keep polling until the
// hostname reported matches the server that was supposed to receive
// that address. A mismatch is treated as "not yet" rather than a
// failure, because during the reboot window it genuinely is. Only the
// timeout is fatal.
//
// The same reasoning means a settle delay before the first attempt: the
// reboot is requested through the API and takes a moment to begin, and
// polling into that gap tests the previous state.
func stepVerify(ctx context.Context, c clients, st *state) error {
	if err := requireKey(st, "verify"); err != nil {
		return err
	}
	if err := st.resolveServers(); err != nil {
		return err
	}

	// Ask the API which address each server holds *now*, rather than
	// working out which address should have ended up where. That is
	// correct whether this runs inside the journey — where the state
	// still records the pre-swap addresses — or on its own afterwards,
	// and it removes a crossover that is easy to get backwards.
	type check struct {
		expectHost string
		addr       string
	}
	names := []string{st.nameA, st.nameB}
	checks := make([]check, 0, len(names))
	for _, name := range names {
		if name == "" {
			continue
		}
		time.Sleep(throttle)
		ip, err := primaryIPv4(ctx, c.server, name)
		if err != nil {
			return err
		}
		log.Printf("  %s should now answer on %s", name, ip.IPAddr)
		checks = append(checks, check{expectHost: name, addr: ip.IPAddr})
	}
	if len(checks) == 0 {
		return fmt.Errorf("no servers to verify")
	}

	log.Printf("  letting the reboots begin (%s) before polling", verifySettle)
	time.Sleep(verifySettle)

	for _, chk := range checks {
		log.Printf("  waiting for %s to answer on %s as itself (up to %s)", chk.expectHost, chk.addr, verifyTimeout)
		if err := waitForHost(chk.addr, chk.expectHost, verifyTimeout); err != nil {
			return fmt.Errorf("%s did not come back on %s: %w", chk.expectHost, chk.addr, err)
		}
		log.Printf("✓ %s answers on %s and identifies as itself", chk.expectHost, chk.addr)
	}

	log.Printf("✓ cutover complete: both guests are reachable on their swapped addresses")
	return nil
}

// waitForHost polls until the address is answered by the expected
// machine.
//
// Anything else — refused, timed out, or answered by a different host —
// counts as not-yet, because all three happen legitimately while a
// reboot is in flight.
func waitForHost(addr, expectHost string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var last string
	for time.Now().Before(deadline) {
		out, err := sshRun(addr, "hostname")
		switch {
		case err != nil:
			last = err.Error()
		case strings.EqualFold(strings.TrimSpace(out), expectHost):
			return nil
		case strings.TrimSpace(out) == "":
			// Connected but the command produced nothing — the guest is
			// probably still coming up.
			last = "answered but reported no hostname"
		default:
			last = fmt.Sprintf("answered as %q, waiting for %q", strings.TrimSpace(out), expectHost)
		}
		time.Sleep(10 * time.Second)
	}
	return fmt.Errorf("timed out after %s: %s", timeout, last)
}
