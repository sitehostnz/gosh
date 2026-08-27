package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/sitehostnz/gosh/pkg/api/server/firewall"
	"github.com/sitehostnz/gosh/pkg/api/server/firewall/securitygroups"
)

// noFirewallFragment identifies the response for a product with no
// firewall at all. Security groups are a high-performance feature;
// standard products reject server/firewall/get outright.
const noFirewallFragment = "not available for this server type"

// stepFirewall reports each server's security groups and whether
// inbound SSH looks open.
//
// Check this before you need it. The rescue environment used to repair
// a server's network configuration is reached over the network, so a
// group that drops inbound SSH closes the recovery path and leaves the
// console as the only way in — which is exactly when you discover it.
//
// A server with no groups attached is reported and skipped: an
// unattached firewall has no rules to judge, and a silent pass here
// would read as "SSH is fine" when nothing was examined.
func stepFirewall(ctx context.Context, c clients, st *state) error {
	if err := st.resolveServers(); err != nil {
		return err
	}
	for _, name := range []string{st.nameA, st.nameB} {
		if name == "" {
			continue
		}
		if err := inspectFirewall(ctx, c, name); err != nil {
			return err
		}
	}
	return nil
}

// inspectFirewall reports one server's groups, tolerating products that
// have no firewall at all.
func inspectFirewall(ctx context.Context, c clients, name string) error {
	time.Sleep(throttle)
	resp, err := c.fw.Get(ctx, firewall.GetRequest{ServerName: name})
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), noFirewallFragment) {
			log.Printf("  %s: no firewall on this product — security groups are an HPVS feature", name)
			return nil
		}
		return fmt.Errorf("firewall.Get(%s): %w", name, err)
	}
	if len(resp.Return) == 0 {
		log.Printf("  %s: no security groups attached — nothing to block the rescue path", name)
		return nil
	}
	for _, attached := range resp.Return {
		if err := reportGroup(ctx, c, name, attached.Group); err != nil {
			return err
		}
	}
	return nil
}

// reportGroup says whether one group would let SSH through.
func reportGroup(ctx context.Context, c clients, server, group string) error {
	time.Sleep(throttle)
	got, err := c.sg.Get(ctx, securitygroups.GetRequest{Name: group})
	if err != nil {
		return fmt.Errorf("securitygroups.Get(%s): %w", group, err)
	}
	blocked, considered := sshBlocked(got.Return.Rules.In)
	switch {
	case considered == 0:
		log.Printf("  %s: group %q has no inbound rule covering SSH", server, group)
	case blocked:
		log.Printf("  %s: group %q appears to BLOCK inbound SSH — rescue mode may be unreachable, keep console access", server, group)
	default:
		log.Printf("  %s: group %q permits inbound SSH (%d rule(s) considered)", server, group, considered)
	}
	return nil
}

// sshBlocked inspects inbound rules covering TCP port 22 and reports
// whether the first matching enabled rule denies it, plus how many
// rules were considered. The count lets a caller tell "explicitly
// allowed" from "no rule mentions SSH", rather than reporting a pass
// for a decision never made.
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
