package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sitehostnz/gosh/pkg/api/server"
	"github.com/sitehostnz/gosh/pkg/api/server/firewall/securitygroups"
	"github.com/sitehostnz/gosh/pkg/api/server/snapshot"
)

// stepInventory reads everything about the account that does not need a
// server of its own.
//
// # Why it exists
//
// These eight endpoints were the ones no step called. That is not a
// cosmetic gap: an endpoint nobody calls has no observed behaviour,
// only assumed behaviour, and this SDK has shipped several methods that
// had never once decoded a real response. Two of them — ListUpgrades
// and the security-group listing — were found exactly this way.
//
// Read-only throughout, so it is safe to run against a production
// account, and it needs no opt-in.
//
// # Logging discipline
//
// Counts and shapes only. Server names, labels and addresses are
// customer data and do not belong in an example's output.
func stepInventory(ctx context.Context, c clients, st *state) error {
	reads := []struct {
		what string
		run  func(context.Context, clients, *state) error
	}{
		{"servers", readServers},
		{"addresses", readAddresses},
		{"resources", readResources},
		{"security groups", readSecurityGroups},
		{"snapshots", readSnapshots},
		{"statistics", readStatistics},
	}
	for _, r := range reads {
		if err := r.run(ctx, c, st); err != nil {
			return fmt.Errorf("%s: %w", r.what, err)
		}
	}
	return nil
}

// readServers lists the account's servers.
func readServers(ctx context.Context, c clients, st *state) error {
	time.Sleep(throttle)
	list, err := c.server.List(ctx)
	if err != nil {
		return fmt.Errorf("server.List: %w", err)
	}
	log.Printf("✓ %d server(s) on this account", len(list.Return.Servers))

	// Choose the subject for the per-server reads below, and keep it
	// local.
	//
	// This step must not write to shared state. It ran before
	// provisioning in the journey, set state.nameA to the first server
	// on the account, and a later step then attached a security group
	// to somebody's real server instead of the one the journey had
	// made. A read-only step that quietly redirects the write steps
	// after it is worse than one that does nothing.
	st.subject = os.Getenv("SH_SERVER_A")
	if st.subject == "" && len(list.Return.Servers) > 0 {
		st.subject = list.Return.Servers[0].Name
		log.Printf("  no SH_SERVER_A; using the first listed for the per-server reads")
	}

	var missing int
	for _, s := range list.Return.Servers {
		if s.Name == "" || s.ProductType == "" {
			missing++
		}
	}
	if missing > 0 {
		return fmt.Errorf("%d of %d servers decoded without a name or product type; the listing shape has changed",
			missing, len(list.Return.Servers))
	}
	return nil
}

// readAddresses covers the two IP listings, which answer different
// questions and are easy to confuse.
//
// ListAllocatedIPs is what this account already holds. ListIPs is
// what is free to allocate at a location. Neither substitutes for the
// other, and a provision needs the second.
func readAddresses(ctx context.Context, c clients, st *state) error {
	time.Sleep(throttle)
	allocated, err := c.server.ListAllocatedIPs(ctx)
	if err != nil {
		return fmt.Errorf("ListAllocatedIPs: %w", err)
	}
	log.Printf("✓ %d address(es) allocated to this account", len(allocated.Return))

	// The map is keyed by a transformed address — IPv6 colons become
	// dots — so the key is not the literal. IPAddr carries that.
	for key, ip := range allocated.Return {
		if ip.IPAddr == "" {
			return fmt.Errorf("an allocated address decoded with an empty IPAddr under key %q; the key is transformed and is not a usable literal", key)
		}
	}

	time.Sleep(throttle)
	free, err := c.server.ListIPs(ctx, server.ListIPsOptions{Location: st.cfg.location})
	if err != nil {
		return fmt.Errorf("ListIPs: %w", err)
	}
	log.Printf("✓ %d address(es) free to allocate at %s", len(free.Return), st.cfg.location)
	return nil
}

// readResources lists the resource groups the account can draw on.
func readResources(ctx context.Context, c clients, _ *state) error {
	time.Sleep(throttle)
	res, err := c.server.ListResources(ctx)
	if err != nil {
		return fmt.Errorf("ListResources: %w", err)
	}
	log.Printf("✓ %d resource group(s)", len(res.Return))
	return nil
}

// readSecurityGroups lists the account's security groups.
//
// This listing had never decoded before August 2026: servers was
// declared as a list of strings where the API sends objects carrying a
// name and a label, so every call failed. Nothing called it, so nothing
// noticed.
func readSecurityGroups(ctx context.Context, c clients, _ *state) error {
	time.Sleep(throttle)
	groups, err := c.sg.List(ctx, securitygroups.ListAllRequest{})
	if err != nil {
		return fmt.Errorf("securitygroups.List: %w", err)
	}
	log.Printf("✓ %d security group(s)", len(groups.Return.Data))

	var attached int
	for _, g := range groups.Return.Data {
		attached += len(g.Servers)
	}
	if len(groups.Return.Data) > 0 {
		log.Printf("  %d server attachment(s) across them", attached)
	}
	return nil
}

// readSnapshots lists the snapshots of the subject server.
func readSnapshots(ctx context.Context, c clients, st *state) error {
	if st.subject == "" {
		log.Printf("  no server to read snapshots for; skipping")
		return nil
	}
	time.Sleep(throttle)
	snaps, err := c.snap.List(ctx, snapshot.ListOptions{Name: st.subject})
	if err != nil {
		return fmt.Errorf("snapshot.List: %w", err)
	}
	log.Printf("✓ %d snapshot(s) on the subject server", len(snaps.Return))
	return nil
}

// readStatistics enumerates the metric types and reads one back.
//
// Note the parameter name: these two take "server_name" where their
// siblings take plain "name". Getting it wrong is rejected in a way
// that reads like the server is missing rather than the parameter.
func readStatistics(ctx context.Context, c clients, st *state) error {
	if st.subject == "" {
		log.Printf("  no server to read statistics for; skipping")
		return nil
	}

	time.Sleep(throttle)
	types, err := c.server.ListStatisticTypes(ctx, server.ListStatisticTypesOptions{ServerName: st.subject})
	if err != nil {
		return fmt.Errorf("ListStatisticTypes: %w", err)
	}
	log.Printf("✓ %d statistic type(s) available: %v", len(types.Return), types.Return.Names())

	// A loop over an empty list would log nothing and read as a pass.
	if len(types.Return) == 0 {
		log.Printf("  no statistic types; skipping GetStatistics")
		return nil
	}

	// Pick a metric and, where the metric needs one, an item to break
	// it down by. Both come from the listing above rather than being
	// guessed, which is the whole point of calling it first.
	var metric, item string
	for name, params := range types.Return {
		metric = name
		for _, p := range params {
			if p.Partition != "" {
				item = p.Partition
			} else if p.Iface != "" {
				item = p.Iface
			}
		}
		break
	}

	time.Sleep(throttle)
	stats, err := c.server.GetStatistics(ctx, server.GetStatisticsOptions{
		ServerName: st.subject,
		Type:       metric,
		Item:       item,
	})
	if err != nil {
		return fmt.Errorf("GetStatistics(%s): %w", metric, err)
	}
	log.Printf("✓ %s decoded: %d series over the reported window", metric, len(stats.Return))
	return nil
}
