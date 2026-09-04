package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sitehostnz/gosh/pkg/api/dns"
)

// stepDelete removes the zone this run created.
//
// Always last, and always attempted: the journey runs it even when an
// earlier step failed, because a zone left behind is clutter that gets
// forgotten.
//
// It removes only what this process created. A zone named in SH_ZONE is
// not touched — the server journey learned that rule the hard way,
// where naming a server for a read-only step was being read as consent
// to destroy it. Deleting a zone this process did not create takes
// SH_DELETE_ZONE, which exists for no other purpose.
func stepDelete(ctx context.Context, c clients, st *state) error {
	name := st.cfg.zone
	if !st.created {
		named := envOr("SH_DELETE_ZONE", "")
		if named == "" {
			log.Printf("  nothing to delete: this run created no zone")
			return nil
		}
		name = named
		log.Printf("  SH_DELETE_ZONE names a zone this process did not create: %s", name)
	}

	time.Sleep(throttle)
	if _, err := c.dns.DeleteZone(ctx, dns.DeleteZoneRequest{DomainName: name}); err != nil {
		return fmt.Errorf("DeleteZone %s: %w", name, err)
	}

	// Confirm it is gone rather than trusting the status. GetZone is a
	// search, so absence is an empty list — which is exactly what we
	// want here, and the one case where that behaviour is convenient.
	time.Sleep(throttle)
	found, err := c.dns.GetZone(ctx, dns.GetZoneRequest{DomainName: name})
	if err != nil {
		return fmt.Errorf("GetZone after delete: %w", err)
	}
	for _, z := range found.Return {
		if z.Name == name {
			return fmt.Errorf("zone %s still present after DeleteZone reported success", name)
		}
	}
	st.created = false
	log.Printf("✓ deleted zone %s, and confirmed it is gone", name)
	return nil
}
