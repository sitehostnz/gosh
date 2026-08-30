package main

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/sitehostnz/gosh/pkg/api/server"
)

// stepPrestage writes each guest's post-swap network configuration from
// inside the guest, without applying it.
//
// # Why this has to happen before the swap
//
// This is the ordering constraint the API does not tell you about. The
// moment an address moves, the guest that answered on it is
// unreachable — so there is no "log in afterwards and fix it up". The
// only ways in after a swap are the rescue environment or the console,
// both out of band, and the rescue environment is a different image
// with its own credentials, so a key injected at provision time does
// not open it.
//
// Nothing on the platform fixes this for you. cloud-init provisions the
// guest on first boot — user, keys, hostname — and then hands
// networking off to a static file on disk. Changing an address updates
// the hypervisor's view, not the file inside the guest, so the guest
// keeps trying to use an address it no longer has.
//
// So the configuration is written first, while both servers are still
// reachable, and deliberately *not* applied: applying it here would cut
// the connection immediately and strand the server on an address it
// does not own yet. Step 70 reboots each server through the API, which
// needs no access to the guest, so there is no window to race.
//
// # Where the configuration comes from, and the one edit it needs
//
// Mostly not from this program. generate_network_config returns the
// exact file the platform would write for a server — the addresses, the
// gateway, the region's resolvers, and the MAC the interface is matched
// on. Rendering that YAML by hand looks easy and is wrong, because the
// interface is matched on a MAC rather than a name.
//
// Which server to ask is the interesting part. After the swap, A holds
// what is currently B's address, so B's current file has the addressing
// A will need. But it also has *B's* MAC — and the MAC does not move
// with the address. Verified live: after a swap the address arrives on
// a server whose interface keeps its own MAC. So neither file is
// correct as-is:
//
//	A's own file     right MAC, stale address
//	B's current file right address, wrong MAC
//
// So this step takes the other server's file and substitutes the
// receiving server's own MAC into it — one well-defined edit, asserted
// to have happened, rather than rendering the whole thing. Everything
// else in the file is the platform's own output.
func stepPrestage(ctx context.Context, c clients, st *state) error {
	if err := requireKey(st, "prestage"); err != nil {
		return err
	}
	if err := st.requirePair(ctx, c, "prestage"); err != nil {
		return err
	}

	// Fetch each server's current configuration; each is what the other
	// server will need once the addresses move.
	configFor := map[string]map[string]string{}
	for _, name := range []string{st.nameA, st.nameB} {
		time.Sleep(throttle)
		resp, err := c.server.GenerateNetworkConfig(ctx, server.GenerateNetworkConfigOptions{Name: name})
		if err != nil {
			return fmt.Errorf("GenerateNetworkConfig(%s): %w", name, err)
		}
		if len(resp.Return) == 0 {
			return fmt.Errorf("GenerateNetworkConfig(%s): no files returned", name)
		}
		configFor[name] = resp.Return
	}

	// Cross them over: A gets B's config, B gets A's.
	pairs := []struct {
		target string
		addr   string
		from   string
		ownMAC string
	}{
		{st.nameA, st.ipA.IPAddr, st.nameB, st.ipA.MacAddr},
		{st.nameB, st.ipB.IPAddr, st.nameA, st.ipB.MacAddr},
	}

	for _, p := range pairs {
		files, err := retargetMAC(configFor[p.from], p.ownMAC)
		if err != nil {
			return fmt.Errorf("prestage %s: %w", p.target, err)
		}
		log.Printf("  %s: reachable on %s, staging %s's config with %s's own MAC %s (%d file(s))",
			p.target, p.addr, p.from, p.target, p.ownMAC, len(files))
		if err := prestageOne(p.addr, files); err != nil {
			return fmt.Errorf("prestage %s (%s): %w", p.target, p.addr, err)
		}
		log.Printf("✓ %s: staged and validated, not applied", p.target)
	}

	log.Printf("  both guests still answer on their current addresses; nothing has changed yet")
	return nil
}

// prestageOne writes every file in one SSH command, so a dropped
// connection cannot leave the guest half-configured.
//
// Assumes the login account has passwordless sudo, which SiteHost images
// provide. An image without it fails here with a sudo password prompt
// rather than anything more informative.
//
// Note what is absent: no netplan apply and no reboot. Either would
// strand the server on an address it does not own yet.
func prestageOne(addr string, files map[string]string) error {
	// Deterministic order, so a failed run is reproducible.
	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	var b strings.Builder
	b.WriteString("set -e\n")
	// Stop cloud-init reasserting its own idea of the network on boot.
	b.WriteString("printf 'network: {config: disabled}\\n' | sudo tee /etc/cloud/cloud.cfg.d/99-disable-network-config.cfg >/dev/null\n")
	for _, path := range paths {
		body := files[path]
		if strings.Contains(body, "PRESTAGE_EOF") {
			return fmt.Errorf("config for %s contains the heredoc delimiter", path)
		}
		fmt.Fprintf(&b, "sudo install -d -m 755 \"$(dirname %s)\"\n", shellQuote(path))
		fmt.Fprintf(&b, "sudo tee %s >/dev/null <<'PRESTAGE_EOF'\n%s\nPRESTAGE_EOF\n", shellQuote(path), body)
		if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
			fmt.Fprintf(&b, "sudo chmod 600 %s\n", shellQuote(path))
		}
	}
	// Validate now, while the error is still visible to us.
	b.WriteString("if command -v netplan >/dev/null 2>&1; then sudo netplan generate; fi\n")

	out, err := sshRun(addr, b.String())
	if err != nil {
		return fmt.Errorf("%w (output: %s)", err, out)
	}
	return nil
}

// shellQuote wraps a path in single quotes for safe interpolation.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// retargetMAC rewrites the macaddress the config matches on, so the
// receiving server's interface is the one selected.
//
// The MAC stays with the server's NIC while the address moves, so a
// config lifted wholesale from the other server would match an
// interface that does not exist here and silently fail to apply. This
// substitutes the receiving server's own MAC and asserts the
// substitution actually happened, rather than trusting that the file
// had the shape expected.
func retargetMAC(files map[string]string, ownMAC string) (map[string]string, error) {
	if ownMAC == "" {
		return nil, fmt.Errorf("no MAC known for the receiving server; cannot retarget the config")
	}
	out := make(map[string]string, len(files))
	edited := 0
	for path, body := range files {
		replaced := macLine.ReplaceAllStringFunc(body, func(line string) string {
			edited++
			indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
			return indent + "macaddress: " + ownMAC
		})
		out[path] = replaced
	}
	if edited == 0 {
		return nil, fmt.Errorf("no macaddress line found in %d generated file(s); the config shape has changed and this step needs revisiting", len(files))
	}
	return out, nil
}

// macLine matches a netplan macaddress entry, whatever the indentation
// and quoting.
var macLine = regexp.MustCompile(`(?m)^[ \t]*macaddress:[ \t]*['"]?[0-9a-fA-F:]*['"]?[ \t]*$`)
