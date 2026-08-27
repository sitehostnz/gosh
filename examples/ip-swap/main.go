// Program ip-swap is an assertion-style validation of swapping IPv4
// addresses between two servers — the operation a failover or a
// like-for-like rebuild depends on.
//
// Unlike every other example in this repo it is NOT read-only: it
// provisions two servers, swaps their addresses and deletes them
// again. It therefore refuses to do anything unless
// SH_EXAMPLE_ALLOW_PROVISION=1 is set. Without that variable it prints
// the rules below and exits zero, so it stays safe to run blind.
//
// # The rule, and how it differs by platform
//
// The obvious approach is to take A's address and add it to B while B
// is still running on its own address. Whether that works depends on
// the product family, which is the part nobody documents:
//
//	Standard performance (LINVPS)  refused, when the two servers are
//	                              in different networks
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
// High-performance servers do not enforce that at all. A single HPVS
// server will accept addresses from two different networks at once,
// with different gateways — verified live on servers in
// 223.165.76.0/24 (network 434) and 223.165.77.0/24 (network 363),
// August 2026.
//
// Treat that as a missing validation rather than a feature. It is a
// suspected platform bug, reported separately, so this example does
// not assert it and does not rely on it: asserting current buggy
// behaviour would mean the example starts failing when the bug is
// fixed. On high performance the check is skipped explicitly.
//
// # The sequence that works on both
//
// Release both servers, then cross-assign:
//
//  1. RemoveIP  A, addrA      — A now holds nothing
//  2. RemoveIP  B, addrB      — B now holds nothing
//  3. AddIP     B, addrA      — accepted: B is empty
//  4. AddIP     A, addrB      — accepted: A is empty
//
// This is the order the KB procedure prescribes, and it is required on
// standard performance. On high performance you could move addresses
// one at a time instead; this example uses the release-both sequence
// regardless, because it is correct on both platforms:
// https://kb.sitehost.nz/servers/ip-addresses/swapping-ip-addresses-between-servers
//
// A server is allowed to hold zero addresses in between. Removing a
// server's only, primary address succeeds.
//
// The refusal is only observable between steps 1 and 2, when addrA is
// free but B is still occupied. Attempt it earlier and the address is
// still in use, so a different error fires and the address-space rule
// is never reached. This example probes in that window, and asserts
// the refusal only on standard performance.
//
// Addresses cannot be moved between locations on either platform — "an
// IP address can't easily be moved to a different region" per the KB —
// and this example does not attempt it.
//
// # Telling the two rejections apart
//
// add_ip has a second rejection that is easy to confuse with the
// address-space one:
//
//	"This ip address is currently in use, or you don't have
//	 permission to use it."   → still attached to something, or
//	                            not allocated to this client
//	"Im sorry this address space cannot be used here."
//	                          → the target server is occupied by an
//	                            address from a different network
//
// A caller who tests the rule without first releasing the address sees
// only the first message and learns nothing about subnets.
//
// # models.IP.NetworkID is the network identity
//
// Two addresses belong to the same network when their NetworkID
// matches. Comparing /24s by string happens to agree for the IPv4
// cases seen so far, but NetworkID is the platform's own identifier
// and is what this example asserts on.
//
// # After the swap: reboot, network config, and getting locked out
//
// Swapping addresses in the API does not touch the guests. Each
// address carries its own netmask and gateway, so once the swap lands
// both guests are still configured for the addresses they no longer
// have, and neither is reachable over the network. Someone has to
// rewrite each guest's network configuration and reboot it.
//
// server.GenerateNetworkConfig returns the files to write, keyed by
// path. Getting them onto the guest is out of band, and there are two
// routes:
//
//   - Rescue mode. Boot the server into its rescue environment
//     (server.ChangeState with "rescue_on"), mount the root filesystem
//     at /mnt, write the files, then "rescue_off" and boot normally.
//     This is the route the KB documents.
//   - Console. An operator makes the same edit over the console, which
//     needs no working network on the guest at all.
//
// The order matters for the operator's own access: rescue mode is
// itself reached over the network, so the recovery path can be firewalled
// off. On HPVS servers a security group that blocks inbound SSH can
// leave the rescue environment unreachable, and then the console is
// the only way back in. Check the firewall before you need it, not
// after — this example reports each server's security groups and
// whether inbound SSH looks blocked before it swaps anything.
//
// Security groups are an HPVS feature. On standard VPS products
// server/firewall/get returns "Firewall functionality is not available
// for this server type", and there is no firewall to lock you out of
// rescue mode. The example reports that and carries on.
//
// # Choosing a product and image
//
// This example provisions a high-performance (HPVS) server. Their
// images are not in the default catalogue: they live behind
// server.ImageTypeHPVMDistro plus a mandatory location, and their
// codes carry a build date (ubuntu-2404-20260727) that changes when an
// image is rebuilt. The example therefore discovers the code at run
// time rather than hardcoding one. Set SH_IMAGE to override.
//
// Standard-performance (LINVPS) products still work — set SH_PRODUCT
// and SH_IMAGE together, since a LINVPS product needs a code from the
// default catalogue and rejects a high-performance one. That platform
// is the older of the two, so it is not the default here.
//
// A server's name is also not its label: the platform truncates the
// label, resolves collisions by appending a digit, and returns the
// real name in the provision response. Every later call must use the
// returned Name.
//
// # On logging addresses
//
// Other examples in this repo log counts and shapes only. This one
// logs full IP addresses, because the addresses are the subject of the
// assertions and they belong to servers this program created moments
// earlier and deletes before it exits. No pre-existing account data is
// read or logged.
//
// Required env: SH_API_KEY, SH_CLIENT_ID, SH_EXAMPLE_ALLOW_PROVISION=1.
// Optional env: SH_LOCATION, SH_PRODUCT, SH_IMAGE, SH_DISTRO, SH_BASE_URL.
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/sitehostnz/gosh/pkg/api"
	"github.com/sitehostnz/gosh/pkg/api/job"
	"github.com/sitehostnz/gosh/pkg/api/server"
	"github.com/sitehostnz/gosh/pkg/api/server/firewall"
	"github.com/sitehostnz/gosh/pkg/api/server/firewall/securitygroups"
	"github.com/sitehostnz/gosh/pkg/models"
)

const (
	// throttle spaces out calls. The API rejects bursts with HTTP 500
	// and a "you have exceeded the number of requests per second"
	// message.
	throttle = 1500 * time.Millisecond

	// jobTimeout bounds a single job. Provisioning is the slow one;
	// address moves complete in seconds.
	jobTimeout = 15 * time.Minute

	// addressSpaceFragment identifies the rejection returned when the
	// target server still holds an address from another network.
	// Matched on a fragment because the surrounding punctuation is not
	// worth depending on.
	addressSpaceFragment = "address space cannot be used"

	// productTypeHPVS is the product family reported by server.Get for
	// high-performance products. Standard performance reports LINVPS.
	productTypeHPVS = "HPVS"

	// noFirewallFragment identifies the response for a product that has
	// no firewall at all. Security groups are an HPVS feature; standard
	// VPS products reject server/firewall/get outright.
	noFirewallFragment = "not available for this server type"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("ip-swap: %v", err)
	}
}

// config carries the provisioning parameters, all overridable so the
// example never hardcodes an account's location or product.
type config struct {
	location string
	product  string
	image    string
	distro   string
}

// newConfig reads the provisioning parameters from the environment.
func newConfig() config {
	return config{
		location: envOr("SH_LOCATION", server.LocationAKLNCT),
		product:  envOr("SH_PRODUCT", "LHPVS1"),
		// Image is discovered rather than defaulted: high-performance
		// image codes carry a build date and change when images are
		// rebuilt, so a literal here would rot. SH_IMAGE overrides.
		image:  os.Getenv("SH_IMAGE"),
		distro: envOr("SH_DISTRO", "ubuntu-noble"),
	}
}

func run() error {
	apiKey := os.Getenv("SH_API_KEY")
	clientID := os.Getenv("SH_CLIENT_ID")
	if apiKey == "" || clientID == "" {
		return fmt.Errorf("SH_API_KEY and SH_CLIENT_ID required")
	}

	if os.Getenv("SH_EXAMPLE_ALLOW_PROVISION") != "1" {
		printRules()
		return nil
	}

	cfg := newConfig()

	c, err := api.New(apiKey, clientID)
	if err != nil {
		return fmt.Errorf("api.New: %w", err)
	}
	s := server.New(c)
	j := job.New(c)
	fw := firewall.New(c)
	sg := securitygroups.New(c)
	ctx := context.Background()

	if cfg.image == "" {
		if cfg.image, err = stepDiscoverImage(ctx, s, cfg); err != nil {
			return err
		}
	}

	if err := stepPreflight(ctx, s, cfg); err != nil {
		return err
	}

	// Names of everything provisioned, so teardown runs even when an
	// assertion below fails. A leaked server bills until someone
	// notices.
	var provisioned []string
	defer func() {
		if err := teardown(ctx, s, provisioned); err != nil {
			log.Printf("✗ teardown: %v", err)
		}
	}()

	return runSwap(ctx, clients{s: s, j: j, fw: fw, sg: sg}, cfg, &provisioned)
}

// clients bundles the API clients the swap needs, so the orchestration
// below reads as a sequence of steps rather than a parameter list.
type clients struct {
	s  *server.Client
	j  *job.Client
	fw *firewall.Client
	sg *securitygroups.Client
}

// runSwap provisions the pair, verifies the refusal, performs the swap
// and reports the guest configuration each server now needs.
func runSwap(ctx context.Context, c clients, cfg config, provisioned *[]string) error {
	a, b, err := stepProvisionPair(ctx, c.s, c.j, cfg, provisioned)
	if err != nil {
		return err
	}

	aIP, bIP, err := stepInspectPair(ctx, c.s, a, b)
	if err != nil {
		return err
	}

	productType, err := stepPlatform(ctx, c.s, a)
	if err != nil {
		return err
	}

	sameNetwork := aIP.NetworkID == bIP.NetworkID
	log.Printf("  topology: %s on network %s (%s), %s on network %s (%s) — %s",
		a, aIP.NetworkID, subnet(aIP.IPAddr),
		b, bIP.NetworkID, subnet(bIP.IPAddr),
		map[bool]string{true: "same network", false: "different networks"}[sameNetwork])

	if err := stepFirewallCheck(ctx, c.fw, c.sg, a, b); err != nil {
		return err
	}

	if err := stepSwap(ctx, c.s, c.j, a, b, aIP, bIP, sameNetwork, productType); err != nil {
		return err
	}

	return stepNetworkConfig(ctx, c.s, a, b)
}

// printRules describes what the example would do, for the default
// no-provisioning path.
func printRules() {
	log.Printf("SH_EXAMPLE_ALLOW_PROVISION is not set — not provisioning anything.")
	log.Printf("")
	log.Printf("This example provisions two servers and swaps their IPv4 addresses.")
	log.Printf("")
	log.Printf("The trap: adding A's address to B while B still holds its own")
	log.Printf("address fails across subnets with")
	log.Printf("    %q", "Im sorry this address space cannot be used here.")
	log.Printf("which reads as \"addresses cannot cross subnets\". They can.")
	log.Printf("")
	log.Printf("The constraint is on the target server's existing addresses, not")
	log.Printf("on the address itself. Release both servers first and the swap")
	log.Printf("succeeds:")
	log.Printf("    1. RemoveIP A, addrA      3. AddIP B, addrA")
	log.Printf("    2. RemoveIP B, addrB      4. AddIP A, addrB")
	log.Printf("")
	log.Printf("Afterwards each guest still needs its network config rewritten;")
	log.Printf("server.GenerateNetworkConfig returns those files.")
	log.Printf("")
	log.Printf("Re-run with SH_EXAMPLE_ALLOW_PROVISION=1 to verify this live.")
}

// stepDiscoverImage picks a high-performance image for the location.
//
// HPVS image codes carry a build date (ubuntu-2404-20260727) and are
// scoped to a location, so they cannot be hardcoded. Requires the
// filters added in #62.
func stepDiscoverImage(ctx context.Context, s *server.Client, cfg config) (string, error) {
	resp, err := s.ListImages(ctx, server.ListImagesOptions{
		Type:     server.ImageTypeHPVMDistro,
		Location: cfg.location,
	})
	if err != nil {
		return "", fmt.Errorf("ListImages: %w", err)
	}
	if len(resp.Return) == 0 {
		return "", fmt.Errorf("ListImages: no %s images at %s", server.ImageTypeHPVMDistro, cfg.location)
	}
	for _, im := range resp.Return {
		// Rows with an empty code exist in the catalogue; skip them.
		if im.Code == "" {
			continue
		}
		if im.Distro == cfg.distro {
			log.Printf("✓ ListImages: %s at %s -> %s (%s)", cfg.distro, cfg.location, im.Code, im.Name)
			return im.Code, nil
		}
	}
	return "", fmt.Errorf("ListImages: no image for distro %q at %s (%d candidates)", cfg.distro, cfg.location, len(resp.Return))
}

// stepPreflight checks the product is available at the location.
//
// CanProvision does not validate the image: it returns true for images
// that provision then rejects as unknown. It is a capacity check, not
// a request check.
func stepPreflight(ctx context.Context, s *server.Client, cfg config) error {
	resp, err := s.CanProvision(ctx, server.CanProvisionOptions{
		Product:  cfg.product,
		Location: cfg.location,
		Distro:   cfg.distro,
	})
	if err != nil {
		return fmt.Errorf("CanProvision: %w", err)
	}
	if !resp.Status {
		return fmt.Errorf("CanProvision: %s/%s unavailable: %s", cfg.product, cfg.location, resp.Msg)
	}
	log.Printf("✓ CanProvision: %s available at %s", cfg.product, cfg.location)
	return nil
}

// stepProvisionPair creates the two servers and waits for both to
// finish building. Each name is recorded as soon as it is known, so a
// failure part-way still tears down what exists.
func stepProvisionPair(
	ctx context.Context, s *server.Client, j *job.Client, cfg config, provisioned *[]string,
) (string, string, error) {
	names := make([]string, 0, 2)
	for _, label := range []string{"gosh-ipswap-a", "gosh-ipswap-b"} {
		time.Sleep(throttle)
		resp, err := s.Create(ctx, server.CreateRequest{
			Label:       label,
			Location:    cfg.location,
			ProductCode: cfg.product,
			Image:       cfg.image,
		})
		if err != nil {
			return "", "", fmt.Errorf("Create(%s): %w", label, err)
		}
		if resp.Return.Name == "" {
			return "", "", fmt.Errorf("Create(%s): API returned no server name", label)
		}
		// The name is not the label — record what the API actually
		// called it, or every later call targets the wrong server.
		name := resp.Return.Name
		*provisioned = append(*provisioned, name)
		names = append(names, name)
		log.Printf("✓ Create: label=%s -> name=%s ips=%v", label, name, resp.Return.Ips)

		if err := waitJob(ctx, j, resp.Return.Job); err != nil {
			return "", "", fmt.Errorf("Create(%s): %w", label, err)
		}
	}
	if len(names) != 2 {
		return "", "", fmt.Errorf("expected 2 servers, got %d", len(names))
	}
	return names[0], names[1], nil
}

// stepPlatform reports the product family, which decides whether the
// address-space constraint applies.
func stepPlatform(ctx context.Context, s *server.Client, name string) (string, error) {
	time.Sleep(throttle)
	resp, err := s.Get(ctx, server.GetRequest{ServerName: name})
	if err != nil {
		return "", fmt.Errorf("Get(%s): %w", name, err)
	}
	if resp.Server.ProductType == "" {
		return "", fmt.Errorf("Get(%s): server reports no product type", name)
	}
	log.Printf("✓ platform: %s (%s)", resp.Server.ProductType, resp.Server.ProductCode)
	return resp.Server.ProductType, nil
}

// stepInspectPair reads both servers back and returns their primary
// IPv4 addresses, asserting each server actually has one.
func stepInspectPair(
	ctx context.Context, s *server.Client, a, b string,
) (models.IP, models.IP, error) {
	var out [2]models.IP
	for i, name := range []string{a, b} {
		time.Sleep(throttle)
		ip, err := primaryIPv4(ctx, s, name)
		if err != nil {
			return out[0], out[1], err
		}
		log.Printf("✓ %s: primary %s network=%s gateway=%s", name, ip.IPAddr, ip.NetworkID, ip.Gateway)
		out[i] = ip
	}
	if out[0].IPAddr == out[1].IPAddr {
		return out[0], out[1], fmt.Errorf("both servers report the same primary %s", out[0].IPAddr)
	}
	return out[0], out[1], nil
}

// stepFirewallCheck reports the security groups attached to each
// server and whether inbound SSH appears to be blocked.
//
// This matters before a swap rather than after: the rescue environment
// used to fix the guest's network configuration is reached over the
// network, so a group that drops inbound SSH can close the recovery
// path and leave console access as the only way in.
//
// A server with no groups attached is reported as such and skipped
// explicitly — an unattached firewall has no rules to judge, and a
// silent pass here would read as "SSH is fine" when nothing was
// checked.
func stepFirewallCheck(
	ctx context.Context, fw *firewall.Client, sg *securitygroups.Client, names ...string,
) error {
	for _, name := range names {
		time.Sleep(throttle)
		resp, err := fw.Get(ctx, firewall.GetRequest{ServerName: name})
		if err != nil {
			// Standard VPS products have no firewall to inspect. That is
			// a property of the product, not a failure, so say so and
			// move on rather than failing a swap that is unaffected.
			if strings.Contains(strings.ToLower(err.Error()), noFirewallFragment) {
				log.Printf("  %s: no firewall on this product — security groups are an HPVS feature", name)
				continue
			}
			return fmt.Errorf("firewall.Get(%s): %w", name, err)
		}
		if len(resp.Return) == 0 {
			log.Printf("  %s: no security groups attached — nothing to block the rescue path", name)
			continue
		}

		for _, attached := range resp.Return {
			time.Sleep(throttle)
			group, err := sg.Get(ctx, securitygroups.GetRequest{Name: attached.Group})
			if err != nil {
				return fmt.Errorf("securitygroups.Get(%s): %w", attached.Group, err)
			}
			blocked, considered := sshBlocked(group.Return.Rules.In)
			switch {
			case considered == 0:
				log.Printf("  %s: group %q has no inbound rules covering SSH", name, attached.Group)
			case blocked:
				log.Printf("  %s: group %q appears to BLOCK inbound SSH — rescue mode may be unreachable, keep console access available", name, attached.Group)
			default:
				log.Printf("  %s: group %q permits inbound SSH (%d rule(s) considered)", name, attached.Group, considered)
			}
		}
	}
	return nil
}

// sshBlocked inspects inbound rules that cover TCP port 22 and reports
// whether the first matching enabled rule denies it, along with how
// many rules were actually considered. The count lets the caller tell
// "explicitly allowed" apart from "no rule mentions SSH at all",
// rather than reporting a pass for a decision never made.
func sshBlocked(in []securitygroups.Rule) (blocked bool, considered int) {
	for _, r := range in {
		if !r.Enabled {
			continue
		}
		if r.Protocol != "" && !strings.EqualFold(r.Protocol, "tcp") {
			continue
		}
		if r.DestPort != 0 && r.DestPort != 22 {
			continue
		}
		considered++
		if considered == 1 {
			blocked = !strings.EqualFold(r.Action, "accept") && !strings.EqualFold(r.Action, "allow")
		}
	}
	return blocked, considered
}

// stepOccupiedTarget asserts that a standard-performance server
// refuses a free address from another network while it still holds its
// own. That refusal is the whole reason the swap sequence releases both
// servers before assigning either.
//
// On high-performance products the same call is accepted, which looks
// like a missing validation rather than an intended capability. This
// example does not make that call on high performance at all: an
// example is a recommendation, and demonstrating a suspected bug would
// both teach an unsupported path and start failing once it is fixed.
// The step is skipped with a note instead.
//
// When both servers land in one network there is nothing to refuse, so
// the step is skipped rather than passing vacuously.
func stepOccupiedTarget(
	ctx context.Context, s *server.Client, b string, aIP models.IP,
	sameNetwork bool, productType string,
) error {
	switch {
	case sameNetwork:
		log.Printf("  skipped: occupied-target refusal (both servers landed in one network, nothing to refuse)")
		return nil
	case productType == productTypeHPVS:
		log.Printf("  skipped: occupied-target refusal (%s does not enforce it; not exercised here)", productType)
		return nil
	}

	time.Sleep(throttle)
	_, err := s.AddIP(ctx, server.AddIPOptions{Name: b, IP: aIP.IPAddr})
	if err == nil {
		return fmt.Errorf("AddIP(%s, %s): a standard-performance server accepted an address from another network while occupied; the constraint this example documents no longer holds", b, aIP.IPAddr)
	}
	if !strings.Contains(strings.ToLower(err.Error()), addressSpaceFragment) {
		return fmt.Errorf("AddIP(%s, %s): expected an address-space refusal, got: %w", b, aIP.IPAddr, err)
	}
	log.Printf("✓ %s refused %s while still holding its own address, as expected", b, aIP.IPAddr)
	return nil
}

// stepSwap performs the real swap: release both servers, then
// cross-assign. This is the KB's order, and the refusal asserted
// half-way through is the reason for it.
func stepSwap(
	ctx context.Context, s *server.Client, j *job.Client, a, b string, aIP, bIP models.IP,
	sameNetwork bool, productType string,
) error {
	// Release A only. Its address is now free, but B still holds its
	// own — which is the state that triggers the address-space refusal.
	if err := release(ctx, s, j, a, aIP.IPAddr); err != nil {
		return err
	}
	log.Printf("✓ %s released %s and holds nothing", a, aIP.IPAddr)

	if err := stepOccupiedTarget(ctx, s, b, aIP, sameNetwork, productType); err != nil {
		return err
	}

	// Release B as well. Both servers now hold nothing.
	if err := release(ctx, s, j, b, bIP.IPAddr); err != nil {
		return err
	}
	log.Printf("✓ %s released %s and holds nothing", b, bIP.IPAddr)

	// Cross-assign. These are the calls that were refused a moment ago,
	// now accepted because each target is empty.
	if err := assign(ctx, s, j, b, aIP.IPAddr); err != nil {
		return err
	}
	if err := assign(ctx, s, j, a, bIP.IPAddr); err != nil {
		return err
	}

	// Assert the swap actually happened, in both directions.
	time.Sleep(throttle)
	if err := assertHolds(ctx, s, b, aIP.IPAddr); err != nil {
		return err
	}
	time.Sleep(throttle)
	if err := assertHolds(ctx, s, a, bIP.IPAddr); err != nil {
		return err
	}
	log.Printf("✓ swapped: %s now holds %s, %s now holds %s", b, aIP.IPAddr, a, bIP.IPAddr)

	return stepSetPrimary(ctx, s, map[string]string{b: aIP.IPAddr, a: bIP.IPAddr})
}

// stepSetPrimary promotes each server's swapped address to primary and
// checks it took.
//
// A server holding exactly one address reports it as primary already,
// so this is a no-op in the common case — but set_primary_ip is the
// call a caller reaches for after a swap, and asserting it here means
// the example covers it rather than leaving it implied.
func stepSetPrimary(ctx context.Context, s *server.Client, want map[string]string) error {
	for name, addr := range want {
		time.Sleep(throttle)
		resp, err := s.SetPrimaryIP(ctx, server.SetPrimaryIPOptions{Name: name, IP: addr})
		if err != nil {
			return fmt.Errorf("SetPrimaryIP(%s, %s): %w", name, addr, err)
		}
		if resp.Return.IPAddr != addr {
			return fmt.Errorf("SetPrimaryIP(%s): returned %q, want %q", name, resp.Return.IPAddr, addr)
		}
		time.Sleep(throttle)
		ip, err := primaryIPv4(ctx, s, name)
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

// release removes one address from a server and asserts the server is
// left holding nothing, which is what makes the cross-network add
// below legal.
func release(ctx context.Context, s *server.Client, j *job.Client, name, addr string) error {
	time.Sleep(throttle)
	resp, err := s.RemoveIP(ctx, server.RemoveIPOptions{Name: name, IP: addr})
	if err != nil {
		return fmt.Errorf("RemoveIP(%s, %s): %w", name, addr, err)
	}
	if err := waitJob(ctx, j, resp.Return.Job); err != nil {
		return fmt.Errorf("RemoveIP(%s): %w", name, err)
	}
	time.Sleep(throttle)
	count, err := ipCount(ctx, s, name)
	if err != nil {
		return err
	}
	if count != 0 {
		return fmt.Errorf("%s: expected 0 addresses after releasing its only one, got %d", name, count)
	}
	return nil
}

// assign attaches one address to an emptied server.
func assign(ctx context.Context, s *server.Client, j *job.Client, name, addr string) error {
	time.Sleep(throttle)
	resp, err := s.AddIP(ctx, server.AddIPOptions{Name: name, IP: addr})
	if err != nil {
		return fmt.Errorf("AddIP(%s, %s): the swap was refused even against an empty server: %w", name, addr, err)
	}
	if err := waitJob(ctx, j, resp.Return.Job); err != nil {
		return fmt.Errorf("AddIP(%s): %w", name, err)
	}
	return nil
}

// stepNetworkConfig fetches the guest network configuration for both
// servers. The API half of a swap is not the whole job: each address
// brings its own netmask and gateway, and without rewriting these
// files the guest does not come back.
func stepNetworkConfig(ctx context.Context, s *server.Client, names ...string) error {
	for _, name := range names {
		time.Sleep(throttle)
		resp, err := s.GenerateNetworkConfig(ctx, server.GenerateNetworkConfigOptions{Name: name})
		if err != nil {
			return fmt.Errorf("GenerateNetworkConfig(%s): %w", name, err)
		}
		if len(resp.Return) == 0 {
			return fmt.Errorf("GenerateNetworkConfig(%s): no files returned; a swapped server would have nothing to write in rescue mode", name)
		}
		paths := make([]string, 0, len(resp.Return))
		for path, body := range resp.Return {
			if len(body) == 0 {
				return fmt.Errorf("GenerateNetworkConfig(%s): %s is empty", name, path)
			}
			paths = append(paths, path)
		}
		log.Printf("✓ %s: guest network config covers %d file(s): %s", name, len(paths), strings.Join(paths, ", "))
	}
	return nil
}

// teardown deletes everything the run created. It reports every
// failure rather than stopping at the first, because anything left
// behind keeps billing.
func teardown(ctx context.Context, s *server.Client, names []string) error {
	var leaked []string
	for _, name := range names {
		time.Sleep(throttle)
		resp, err := s.Delete(ctx, server.DeleteRequest{Name: name, Force: true})
		switch {
		case err != nil:
			log.Printf("✗ Delete(%s): %v", name, err)
			leaked = append(leaked, name)
		case !resp.Status:
			log.Printf("✗ Delete(%s): %s", name, resp.Msg)
			leaked = append(leaked, name)
		default:
			log.Printf("✓ Delete(%s): scheduled", name)
		}
	}
	if len(leaked) > 0 {
		return fmt.Errorf("these servers were NOT deleted and are still billing: %s", strings.Join(leaked, ", "))
	}
	return nil
}

// primaryIPv4 returns the server's primary IPv4 address.
func primaryIPv4(ctx context.Context, s *server.Client, name string) (models.IP, error) {
	resp, err := s.Get(ctx, server.GetRequest{ServerName: name})
	if err != nil {
		return models.IP{}, fmt.Errorf("Get(%s): %w", name, err)
	}
	if len(resp.Server.Ips) == 0 {
		return models.IP{}, fmt.Errorf("Get(%s): server has no addresses", name)
	}
	for _, ip := range resp.Server.Ips {
		if ip.Primary && ip.AddrFamily == 4 {
			return ip, nil
		}
	}
	return models.IP{}, fmt.Errorf("Get(%s): no primary IPv4 among %d address(es)", name, len(resp.Server.Ips))
}

// ipCount reports how many addresses a server currently holds.
func ipCount(ctx context.Context, s *server.Client, name string) (int, error) {
	resp, err := s.Get(ctx, server.GetRequest{ServerName: name})
	if err != nil {
		return 0, fmt.Errorf("Get(%s): %w", name, err)
	}
	return len(resp.Server.Ips), nil
}

// assertHolds fails unless the named server currently holds addr.
func assertHolds(ctx context.Context, s *server.Client, name, addr string) error {
	resp, err := s.Get(ctx, server.GetRequest{ServerName: name})
	if err != nil {
		return fmt.Errorf("Get(%s): %w", name, err)
	}
	for _, ip := range resp.Server.Ips {
		if ip.IPAddr == addr {
			return nil
		}
	}
	return fmt.Errorf("%s does not hold %s after the swap (holds %d address(es))", name, addr, len(resp.Server.Ips))
}

// waitJob polls a job to a terminal state. A zero ID means the call
// was synchronous and there is nothing to wait for.
func waitJob(ctx context.Context, j *job.Client, jb models.Job) error {
	if jb.ID == 0 {
		return nil
	}
	deadline := time.Now().Add(jobTimeout)
	for {
		time.Sleep(throttle * 2)
		resp, err := j.Get(ctx, job.GetRequest{ID: jb.ID, Type: jb.Type})
		if err != nil {
			// A poll can fail on the per-second limit; that is not the
			// job failing. Keep trying until the deadline.
			if time.Now().After(deadline) {
				return fmt.Errorf("job %d: %w", jb.ID, err)
			}
			continue
		}
		switch resp.Return.State {
		case "Completed":
			return nil
		case "Failed":
			return fmt.Errorf("job %d failed: %s", jb.ID, resp.Return.Message)
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("job %d: still %s after %s", jb.ID, resp.Return.State, jobTimeout)
		}
	}
}

// subnet labels an address's /24 (IPv4) or /64 (IPv6), for the
// human-readable topology line only. Assertions use NetworkID.
func subnet(addr string) string {
	ip := net.ParseIP(addr)
	if ip == nil {
		return "unparseable"
	}
	if v4 := ip.To4(); v4 != nil {
		return fmt.Sprintf("%d.%d.%d.0/24", v4[0], v4[1], v4[2])
	}
	return ip.Mask(net.CIDRMask(64, 128)).String() + "/64"
}

// envOr returns the environment variable or a fallback.
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
