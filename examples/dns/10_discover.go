package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sitehostnz/gosh/pkg/api/dns"
)

// stepDiscover reads the account's zones and reverse-DNS addresses.
//
// Read-only, and safe to run anywhere. It is also the step that proves
// the credentials work before any later step tries to change anything.
func stepDiscover(ctx context.Context, c clients, st *state) error {
	time.Sleep(throttle)
	zones, err := c.dns.ListZones(ctx, &dns.ListZoneOptions{})
	if err != nil {
		return fmt.Errorf("ListZones: %w", err)
	}
	log.Printf("✓ %d zone(s) on this account", len(zones.Return.Data))

	// Counts and shapes only — a zone name is customer data.
	if len(zones.Return.Data) == 0 {
		log.Printf("  no zones; nothing to read records from")
	} else {
		var named int
		for _, z := range zones.Return.Data {
			if z.Name != "" {
				named++
			}
		}
		if named != len(zones.Return.Data) {
			return fmt.Errorf("%d of %d zones decoded without a name; the listing shape has changed",
				len(zones.Return.Data)-named, len(zones.Return.Data))
		}
		log.Printf("  every zone decoded with a name")
	}

	// Whether this zone already exists decides what the later steps may
	// do. Checking here means the journey fails early and clearly
	// rather than half-way through creating something.
	for _, z := range zones.Return.Data {
		if z.Name == st.cfg.zone {
			log.Printf("  note: %s already exists; the zone step will refuse to touch it", st.cfg.zone)
		}
	}

	time.Sleep(throttle)
	ips, err := c.dns.ListIPs(ctx)
	if err != nil {
		return fmt.Errorf("ListIPs: %w", err)
	}
	log.Printf("✓ %d address(es) available for reverse DNS", len(ips.Return))
	return nil
}
