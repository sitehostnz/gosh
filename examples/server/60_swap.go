package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/sitehostnz/gosh/pkg/api/job"
	"github.com/sitehostnz/gosh/pkg/api/server"
	"github.com/sitehostnz/gosh/pkg/models"
)

// addressSpaceFragment identifies the refusal returned when a
// standard-performance target still holds an address from another
// network.
const addressSpaceFragment = "address space cannot be used"

// stepSwap exchanges the two servers' primary addresses.
//
// # The rule, and how it differs by platform
//
// The obvious approach is to take A's address and add it to B while B
// still holds its own. Whether that works depends on the product
// family, which is the part nobody documents:
//
//	Standard performance (LINVPS)  refused, when the two servers are
//	                               in different networks
//	High performance (HPVS)        accepted (suspected bug — see below)
//
// On standard performance the refusal is:
//
//	Im sorry this address space cannot be used here.
//
// which reads as "addresses cannot cross subnets". They can. The
// constraint is not on the address, it is on the *target server's
// existing addresses*: a standard-performance server will not accept an
// address from one network while it still holds an address of the same
// family from a different one. Release the target first and the same
// call succeeds.
//
// High-performance servers do not enforce that. An HPVS server will
// accept addresses from two different networks at once, with different
// gateways. Treat that as a missing validation rather than a feature:
// it is a suspected platform bug, reported separately, so this step
// does not exercise it on high performance and does not rely on it.
// Asserting current buggy behaviour would mean this example starts
// failing when the bug is fixed.
//
// # The sequence that works on both
//
//  1. RemoveIP  A, addrA      — A now holds nothing
//  2. RemoveIP  B, addrB      — B now holds nothing
//  3. AddIP     B, addrA      — accepted: B is empty
//  4. AddIP     A, addrB      — accepted: A is empty
//
// A server is allowed to hold zero addresses in between; removing its
// only, primary address succeeds.
//
// The refusal is only observable between steps 1 and 2, while addrA is
// free but B is still occupied. Probe it before step 1 and the address
// is still in use, so a different error fires — "This ip address is
// currently in use, or you don't have permission to use it." — and the
// address-space rule is never reached. That distinction is why a caller
// testing this by hand usually learns nothing.
//
// Addresses cannot be moved between locations on either platform.
func stepSwap(ctx context.Context, c clients, st *state) error {
	if err := st.requirePair(ctx, c, "swap"); err != nil {
		return err
	}

	sameNetwork := st.ipA.NetworkID == st.ipB.NetworkID

	// Release A only. Its address is now free, but B still holds its
	// own — the state that triggers the refusal on standard performance.
	if err := release(ctx, c, st.nameA, st.ipA.IPAddr); err != nil {
		return err
	}
	log.Printf("✓ %s released %s and holds nothing", st.nameA, st.ipA.IPAddr)

	if err := assertOccupiedTarget(ctx, c, st.nameB, st.ipA, sameNetwork, st.productType); err != nil {
		return err
	}

	if err := release(ctx, c, st.nameB, st.ipB.IPAddr); err != nil {
		return err
	}
	log.Printf("✓ %s released %s and holds nothing", st.nameB, st.ipB.IPAddr)

	// Cross-assign: the calls refused a moment ago, now accepted
	// because each target is empty.
	if err := assign(ctx, c, st.nameB, st.ipA.IPAddr); err != nil {
		return err
	}
	if err := assign(ctx, c, st.nameA, st.ipB.IPAddr); err != nil {
		return err
	}

	time.Sleep(throttle)
	if err := assertHolds(ctx, c.server, st.nameB, st.ipA.IPAddr); err != nil {
		return err
	}
	time.Sleep(throttle)
	if err := assertHolds(ctx, c.server, st.nameA, st.ipB.IPAddr); err != nil {
		return err
	}
	log.Printf("✓ swapped: %s holds %s, %s holds %s", st.nameB, st.ipA.IPAddr, st.nameA, st.ipB.IPAddr)

	if err := reportMACMovement(ctx, c, st); err != nil {
		return err
	}

	return setPrimaries(ctx, c, map[string]string{
		st.nameB: st.ipA.IPAddr,
		st.nameA: st.ipB.IPAddr,
	})
}

// assertOccupiedTarget checks that a standard-performance server
// refuses a free address from another network while it still holds its
// own.
//
// On high performance the same call is accepted, which looks like a
// missing validation. This does not make that call there: an example is
// a recommendation, and demonstrating a suspected bug would teach an
// unsupported path. It skips with a note instead.
//
// When both servers share a network there is nothing to refuse, so the
// check is skipped rather than passing vacuously.
func assertOccupiedTarget(
	ctx context.Context, c clients, target string, addr models.IP, sameNetwork bool, productType string,
) error {
	switch {
	case sameNetwork:
		log.Printf("  skipped: occupied-target refusal (both servers are in one network, nothing to refuse)")
		return nil
	case productType == productTypeHPVS:
		log.Printf("  skipped: occupied-target refusal (%s does not enforce it; not exercised here)", productType)
		return nil
	}

	time.Sleep(throttle)
	_, err := c.server.AddIP(ctx, server.AddIPOptions{Name: target, IP: addr.IPAddr})
	if err == nil {
		return fmt.Errorf("AddIP(%s, %s): a standard-performance server accepted an address from another network while occupied; the constraint this example documents no longer holds", target, addr.IPAddr)
	}
	if !strings.Contains(strings.ToLower(err.Error()), addressSpaceFragment) {
		return fmt.Errorf("AddIP(%s, %s): expected an address-space refusal, got: %w", target, addr.IPAddr, err)
	}
	log.Printf("✓ %s refused %s while still holding its own address, as expected", target, addr.IPAddr)
	return nil
}

// release removes one address and asserts the server is left holding
// nothing, which is what makes the cross-network add legal.
func release(ctx context.Context, c clients, name, addr string) error {
	time.Sleep(throttle)
	resp, err := c.server.RemoveIP(ctx, server.RemoveIPOptions{Name: name, IP: addr})
	if err != nil {
		return fmt.Errorf("RemoveIP(%s, %s): %w", name, addr, err)
	}
	if err := waitJob(ctx, c.job, resp.Return.Job); err != nil {
		return fmt.Errorf("RemoveIP(%s): %w", name, err)
	}
	time.Sleep(throttle)
	count, err := ipCount(ctx, c.server, name)
	if err != nil {
		return err
	}
	if count != 0 {
		return fmt.Errorf("%s: expected 0 addresses after releasing its only one, got %d", name, count)
	}
	return nil
}

// assign attaches one address to an emptied server.
func assign(ctx context.Context, c clients, name, addr string) error {
	time.Sleep(throttle)
	resp, err := c.server.AddIP(ctx, server.AddIPOptions{Name: name, IP: addr})
	if err != nil {
		return fmt.Errorf("AddIP(%s, %s): refused even against an empty server: %w", name, addr, err)
	}
	if err := waitJob(ctx, c.job, resp.Return.Job); err != nil {
		return fmt.Errorf("AddIP(%s): %w", name, err)
	}
	return nil
}

// setPrimaries promotes each swapped address and checks it took.
//
// A server holding exactly one address reports it as primary already,
// so this is usually a no-op — but set_primary_ip is the call a caller
// reaches for after a swap, so it is exercised rather than implied.
func setPrimaries(ctx context.Context, c clients, want map[string]string) error {
	for name, addr := range want {
		time.Sleep(throttle)
		resp, err := c.server.SetPrimaryIP(ctx, server.SetPrimaryIPOptions{Name: name, IP: addr})
		if err != nil {
			return fmt.Errorf("SetPrimaryIP(%s, %s): %w", name, addr, err)
		}
		if resp.Return.IPAddr != addr {
			return fmt.Errorf("SetPrimaryIP(%s): returned %q, want %q", name, resp.Return.IPAddr, addr)
		}
		time.Sleep(throttle)
		ip, err := primaryIPv4(ctx, c.server, name)
		if err != nil {
			return err
		}
		if ip.IPAddr != addr {
			return fmt.Errorf("%s: primary is %s, want %s", name, ip.IPAddr, addr)
		}
		log.Printf("✓ %s: primary is %s", name, addr)
	}
	return nil
}

// unused keeps the job import referenced if the file is trimmed.
var _ = job.SchedulerType

// reportMACMovement records whether an address's MAC followed it to the
// new server.
//
// This decides whether staging the other server's generated
// configuration is sound. The generated netplan for a high-performance
// server matches its interface on a MAC address taken from the address
// row, so:
//
//   - If the MAC travels with the address, the configuration generated
//     for B today is exactly what A needs afterwards, and staging it
//     crosswise is correct.
//   - If the MAC stays with the server's NIC, that configuration
//     matches an interface the receiving server does not have, and the
//     staged file will not apply.
//
// Reported rather than asserted: this is an observation about platform
// behaviour, and the useful thing is to see it in the log when a run
// misbehaves. Step 80 is what actually fails if the guests do not come
// back.
func reportMACMovement(ctx context.Context, c clients, st *state) error {
	time.Sleep(throttle)
	nowB, err := primaryIPv4(ctx, c.server, st.nameB)
	if err != nil {
		return err
	}
	switch nowB.MacAddr {
	case "":
		log.Printf("  MAC: not reported for %s after the swap", nowB.IPAddr)
	case st.ipA.MacAddr:
		log.Printf("  MAC: %s travelled with %s (was on %s, now on %s) — staged config is correct",
			nowB.MacAddr, nowB.IPAddr, st.nameA, st.nameB)
	default:
		log.Printf("  MAC: %s now has mac=%s but the address arrived carrying mac=%s",
			nowB.IPAddr, nowB.MacAddr, st.ipA.MacAddr)
		log.Printf("    the MAC did NOT travel with the address; a config matched on the old MAC will not apply")
	}
	return nil
}
