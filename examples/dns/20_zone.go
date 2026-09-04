package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sitehostnz/gosh/pkg/api/dns"
)

// stepZone creates the DNS zone the later steps operate on.
//
// # Creating a zone is not registering a domain
//
// These are separate operations that both get described as "adding a
// domain", and only one of them costs money. This creates a zone in
// SiteHost's DNS for a name, which is free and has no effect unless
// something delegates to those nameservers. Registration is a
// different system entirely.
//
// The default name is under .invalid, reserved by RFC 2606 so that it
// can never be registered by anyone. That makes the journey safe to
// run repeatedly: the zone cannot collide with a real one, and nothing
// resolves it.
//
// # It refuses to adopt a zone it did not create
//
// If the zone already exists, this fails rather than continuing. The
// later steps add and delete records, and doing that to a zone
// somebody else made — because SH_ZONE was pointed at a real one — is
// not recoverable. Cleanup likewise removes only what this step
// created.
func stepZone(ctx context.Context, c clients, st *state) error {
	if err := refuseExisting(ctx, c, st.cfg.zone); err != nil {
		return err
	}

	time.Sleep(throttle)
	created, err := c.dns.CreateZone(ctx, dns.CreateZoneRequest{DomainName: st.cfg.zone})
	if err != nil {
		return fmt.Errorf("CreateZone: %w", err)
	}
	st.created = true
	log.Printf("✓ created zone %s", st.cfg.zone)
	if created.Msg != "" {
		log.Printf("  API said: %s", created.Msg)
	}

	// Read it back. A create that reports success and produces nothing
	// is the failure this catches, and it is the reason the assertion
	// is here rather than assumed.
	tmpl, err := zoneTemplate(ctx, c, st.cfg.zone)
	if err != nil {
		return fmt.Errorf("CreateZone reported success but %s does not read back: %w", st.cfg.zone, err)
	}
	st.originalTemplate = tmpl
	log.Printf("✓ zone reads back, linked to template %q", tmpl)

	// A new zone is not empty: the platform seeds it from a template.
	// Worth reporting, because a caller adding records needs to know
	// what is already there.
	time.Sleep(throttle)
	records, err := c.dns.ListRecords(ctx, dns.ListRecordsRequest{Domain: st.cfg.zone})
	if err != nil {
		return fmt.Errorf("ListRecords: %w", err)
	}
	log.Printf("✓ the new zone starts with %d record(s), seeded from its template", len(records.Return))
	for _, r := range records.Return {
		log.Printf("    %-6s %s", r.Type, r.Name)
	}
	return nil
}

// refuseExisting fails when the zone is already there.
//
// GetZone is a search: absence comes back as an empty list with
// status:true, never an error. Checking err alone here would report
// every zone as already existing.
func refuseExisting(ctx context.Context, c clients, zone string) error {
	time.Sleep(throttle)
	existing, err := c.dns.GetZone(ctx, dns.GetZoneRequest{DomainName: zone})
	if err != nil {
		return fmt.Errorf("GetZone: %w", err)
	}
	for _, z := range existing.Return {
		if z.Name == zone {
			return fmt.Errorf("zone %s already exists; this journey will not add records to a zone it did not create — pick another SH_ZONE, or remove that one first", zone)
		}
	}
	return nil
}
