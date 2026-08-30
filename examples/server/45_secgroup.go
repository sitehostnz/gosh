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

// stepSecGroup walks a security group through its whole life: create,
// attach to a server, change its rules, detach, delete.
//
// # Why this is worth a step of its own
//
// Nothing exercised Add, Update or the listing, and the listing turned
// out never to have decoded — servers was declared as a list of strings
// where the API sends objects. An endpoint nobody calls has no observed
// behaviour, only assumed behaviour.
//
// # It is a high-performance feature
//
// Security groups exist for HPVS products. Legacy Xen (LINVPS) servers
// have no firewall endpoints at all and reject them outright; the
// standard-performance (SVS) tier was not tested. The step reports and
// skips rather than failing when the product has no firewall, because
// that is a property of the server rather than a fault.
//
// # The rules it writes are deliberately permissive
//
// It opens inbound SSH and nothing else. A group that drops SSH would
// close the path the later steps use to reach the guests, and the
// rescue environment is reached the same way — so a firewall mistake
// here would cost a console session to undo. The group is detached and
// deleted before the step returns either way.
func stepSecGroup(ctx context.Context, c clients, st *state) error {
	// Act on a server this run provisioned. The step attaches a
	// firewall and rewrites its rules, which is not something to do to
	// a server somebody else made.
	if len(st.created) == 0 {
		return fmt.Errorf("the secgroup step attaches a firewall; run the journey so it acts on a server this process created")
	}
	name := st.created[0]

	// Confirm the product has a firewall before creating anything, so
	// a Xen server does not leave an orphaned group behind.
	time.Sleep(throttle)
	if _, err := c.fw.Get(ctx, firewall.GetRequest{ServerName: name}); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), noFirewallFragment) {
			log.Printf("  %s has no firewall endpoints; security groups are a high-performance feature, skipping", name)
			return nil
		}
		return fmt.Errorf("firewall.Get: %w", err)
	}

	// The address to probe. Everything below is checked against a real
	// socket, not against the API's account of itself.
	if err := st.ensureAddresses(ctx, c); err != nil {
		return err
	}
	addr := st.ipA.IPAddr

	// Establish the before state. Without it, a "blocked" result later
	// proves nothing — the port might never have been open.
	log.Printf("  before any firewall: is %s:22 reachable?", addr)
	if err := assertReachable(addr, "22", "the journey needs SSH open before it can prove a firewall closed it"); err != nil {
		return err
	}

	label := "gosh-example-" + name
	group, err := createGroup(ctx, c, label)
	if err != nil {
		return err
	}

	// Whatever happens next, take the group off the server and remove
	// it. A leftover group is account clutter, and one still attached
	// is a firewall nobody meant to leave in place.
	defer func() {
		if err := detachAndDelete(ctx, c, name, group); err != nil {
			log.Printf("! could not clean up security group %s: %v", group, err)
			log.Printf("!   remove it by hand, and check %s has no groups attached", name)
		}
	}()

	return proveItFilters(ctx, c, name, group, label, addr)
}

// proveItFilters is the part that does not take the API's word for it.
//
// Drop tcp/22, check the port stops answering; accept it again, check
// it comes back. Both directions matter: a port that is closed the
// whole time would satisfy the first check on its own, and would prove
// nothing about the firewall.
func proveItFilters(ctx context.Context, c clients, name, group, label, addr string) error {
	// Write a rule that drops SSH, then attach. Order matters:
	// attaching an empty group first would prove nothing, since an
	// empty group has nothing to enforce.
	if err := setRules(ctx, c, group, label, "DROP"); err != nil {
		return err
	}
	if err := attachGroup(ctx, c, name, group); err != nil {
		return err
	}

	// The real-world assertion. The API has told us the group is
	// attached and its job completed; this asks the server.
	log.Printf("  a DROP rule for 22 is attached; is %s:22 still reachable?", addr)
	if err := assertBlocked(addr, "22", "the security group drops inbound tcp/22"); err != nil {
		return err
	}

	// And back the other way, so the result is not an accident of
	// something else being down.
	if err := setRules(ctx, c, group, label, "ACCEPT"); err != nil {
		return err
	}
	log.Printf("  the rule is now ACCEPT; is %s:22 reachable again?", addr)
	if err := assertReachable(addr, "22", "the security group now accepts inbound tcp/22"); err != nil {
		return err
	}

	log.Printf("✓ the firewall was observed to change what reaches the server, not merely what the API reports")
	return nil
}

// createGroup adds a group and returns its generated name.
//
// Note the two identifiers. Add takes a Label, which is yours to
// choose, and the API generates a Name — "sge63f6e0daa" — which is
// what every other call wants. Add returns that name directly, so
// there is no need to go looking for it in the listing.
func createGroup(ctx context.Context, c clients, label string) (string, error) {
	time.Sleep(throttle)
	added, err := c.sg.Add(ctx, securitygroups.AddRequest{Label: label})
	if err != nil {
		return "", fmt.Errorf("securitygroups.Add: %w", err)
	}
	if added.Return.Name == "" {
		return "", fmt.Errorf("securitygroups.Add reported success (%s) but returned no name; nothing downstream can address the group", added.Msg)
	}
	log.Printf("✓ created security group %q, named %s by the API", label, added.Return.Name)
	return added.Return.Name, nil
}

// attachGroup puts the group on the server and reads it back.
func attachGroup(ctx context.Context, c clients, name, group string) error {
	// Attaching is asynchronous: the response carries a scheduler job,
	// and reading the server back before that job finishes reports the
	// old groups. That looks exactly like the attach having failed.
	time.Sleep(throttle)
	resp, err := c.fw.Update(ctx, firewall.UpdateRequest{
		ServerName:     name,
		SecurityGroups: []string{group},
	})
	if err != nil {
		return fmt.Errorf("firewall.Update attaching: %w", err)
	}
	if err := waitJob(ctx, c.job, resp.Return.Job); err != nil {
		return fmt.Errorf("firewall.Update attaching: %w", err)
	}

	time.Sleep(throttle)
	got, err := c.fw.Get(ctx, firewall.GetRequest{ServerName: name})
	if err != nil {
		return fmt.Errorf("firewall.Get after attaching: %w", err)
	}
	// The firewall listing reports the group under "group", not "name".
	for _, g := range got.Return {
		if g.Group == group {
			log.Printf("✓ attached %s to %s, and the server reports it", group, name)
			return nil
		}
	}
	return fmt.Errorf("firewall.Update reported success but %s does not report group %s", name, group)
}

// setRules rewrites the group's inbound rules to a single rule for
// tcp/22 with the given action, and reads them back.
//
// The two actions the API accepts are ACCEPT and DROP.
func setRules(ctx context.Context, c clients, group, label, action string) error {
	rules := []securitygroups.UpdateRequestRule{{
		Enabled:         true,
		Action:          action,
		Protocol:        "tcp",
		DestinationPort: "22",
	}}

	// Updating the rules is asynchronous too.
	time.Sleep(throttle)
	resp, err := c.sg.Update(ctx, securitygroups.UpdateRequest{
		Name: group,
		Params: securitygroups.ParamsOptions{
			Label:   label,
			RulesIn: rules,
		},
	})
	if err != nil {
		return fmt.Errorf("securitygroups.Update: %w", err)
	}
	if err := waitJob(ctx, c.job, resp.Return.Job); err != nil {
		return fmt.Errorf("securitygroups.Update: %w", err)
	}

	time.Sleep(throttle)
	got, err := c.sg.Get(ctx, securitygroups.GetRequest{Name: group})
	if err != nil {
		return fmt.Errorf("securitygroups.Get after update: %w", err)
	}
	if len(got.Return.Rules.In) == 0 {
		return fmt.Errorf("securitygroups.Update reported success but %s has no inbound rules", group)
	}
	log.Printf("✓ %s now has %d inbound rule(s), action %s",
		group, len(got.Return.Rules.In), action)
	return nil
}

// detachAndDelete removes the group from the server and deletes it.
func detachAndDelete(ctx context.Context, c clients, name, group string) error {
	time.Sleep(throttle)
	// An empty group list is how a server is detached from everything.
	resp, err := c.fw.Update(ctx, firewall.UpdateRequest{
		ServerName:     name,
		SecurityGroups: []string{},
	})
	if err != nil {
		return fmt.Errorf("firewall.Update detaching: %w", err)
	}
	if err := waitJob(ctx, c.job, resp.Return.Job); err != nil {
		return fmt.Errorf("firewall.Update detaching: %w", err)
	}

	time.Sleep(throttle)
	if _, err := c.sg.Delete(ctx, securitygroups.DeleteRequest{Name: group}); err != nil {
		return fmt.Errorf("securitygroups.Delete: %w", err)
	}
	log.Printf("✓ detached and deleted %s", group)
	return nil
}
