package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sitehostnz/gosh/pkg/api/server"
)

// stepDiscover finds a location and product that can host containers.
//
// Cloud Containers are provisioned through the same server endpoints as
// a virtual server — the product code is what makes it a container —
// so discovery is server.ListLocations plus server.ListProducts
// filtered to the CLDCON family.
//
// Two things worth noticing here, both undocumented:
//
//   - The Linode-backed locations carry Cloud Containers and nothing
//     else. If a location's product types are CLDCON alone, no virtual
//     server product exists there at all.
//   - Cloud Containers do not take an image code. The provision call
//     requires the parameter but does not validate it, so any string is
//     accepted and the platform selects its own image. That is worth
//     knowing before writing code that tries to choose one.
func stepDiscover(ctx context.Context, c clients, st *state) error {
	time.Sleep(throttle)
	locs, err := c.server.ListLocations(ctx)
	if err != nil {
		return fmt.Errorf("ListLocations: %w", err)
	}

	hosting := make([]string, 0, len(locs.Return))
	for _, l := range locs.Return {
		if l.Public != "1" || !carries(l.ProductTypes, "CLDCON") {
			continue
		}
		hosting = append(hosting, l.Code)
		if only := len(l.ProductTypes) == 1; only {
			log.Printf("  %s carries containers and nothing else", l.Code)
		}
	}
	if len(hosting) == 0 {
		return fmt.Errorf("ListLocations: no public location offers CLDCON")
	}
	log.Printf("✓ %d location(s) offer containers: %v", len(hosting), hosting)

	if !carries(hosting, st.cfg.location) {
		return fmt.Errorf("location %s does not offer containers; try one of %v", st.cfg.location, hosting)
	}

	return pickProduct(ctx, c, st)
}

// pickProduct checks the configured product is offered at the
// configured location, and reports what it is.
func pickProduct(ctx context.Context, c clients, st *state) error {
	time.Sleep(throttle)
	prods, err := c.server.ListProducts(ctx, server.ListProductsOptions{
		Location: st.cfg.location,
		Types:    []string{"CLDCON"},
	})
	if err != nil {
		return fmt.Errorf("ListProducts: %w", err)
	}
	if len(prods.Return) == 0 {
		return fmt.Errorf("ListProducts: no CLDCON products at %s", st.cfg.location)
	}
	log.Printf("✓ %d container product(s) at %s", len(prods.Return), st.cfg.location)

	codes := make([]string, 0, len(prods.Return))
	var found *server.Product
	for i, p := range prods.Return {
		codes = append(codes, p.Code)
		if p.Code == st.cfg.product {
			found = &prods.Return[i]
		}
	}
	if found == nil {
		return fmt.Errorf("product %s not offered at %s; available: %v", st.cfg.product, st.cfg.location, codes)
	}
	log.Printf("✓ %s: %d core(s), %.0fGB RAM, %s/month",
		found.Code, found.Attributes.Cores, found.Attributes.RAM, found.Price.String())

	// Containers report no disk attributes, unlike virtual servers.
	if found.Attributes.Disk == 0 && len(found.Attributes.Partitions) == 0 {
		log.Printf("  no disk or partitions reported — containers are sized by cores and RAM")
	}
	return nil
}

// carries reports whether the list holds v.
func carries(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}
