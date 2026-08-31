package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/sitehostnz/gosh/pkg/api/server"
)

// stepDiscover resolves the one input that cannot be hardcoded: the
// image code.
//
// High-performance image codes carry a build date — ubuntu-2404-20260727
// — and change whenever an image is rebuilt, so pinning a literal in an
// example guarantees it breaks later. They also live in a separate
// catalogue from the default listing: reaching them needs
// ImageTypeHPVMDistro *and* a location, and omitting the location is an
// error rather than a wider search.
//
// Legacy Xen (LINVPS) images are in the default catalogue instead, so
// when SH_PRODUCT names a standard product you must supply SH_IMAGE
// too — the two catalogues do not share codes, and a code from one is
// rejected by the other.
func stepDiscover(ctx context.Context, c clients, st *state) error {
	if err := checkLocation(ctx, c, st); err != nil {
		return err
	}
	if st.cfg.image != "" {
		log.Printf("✓ image pinned by SH_IMAGE: %s", st.cfg.image)
	} else if err := resolveImage(ctx, c, st); err != nil {
		return err
	}
	return describeProduct(ctx, c, st)
}

// checkLocation confirms the configured location exists and has
// addresses to hand out.
func checkLocation(ctx context.Context, c clients, st *state) error {
	time.Sleep(throttle)
	locs, err := c.server.ListLocations(ctx)
	if err != nil {
		return fmt.Errorf("ListLocations: %w", err)
	}
	if len(locs.Return) == 0 {
		return fmt.Errorf("ListLocations: no locations returned")
	}
	for _, l := range locs.Return {
		if l.Code != st.cfg.location {
			continue
		}
		log.Printf("✓ location %s (%s): ipv4 available=%d, products=%v",
			l.Code, l.Label, l.AvailableIPv4, l.ProductTypes)
		if l.AvailableIPv4 == 0 {
			return fmt.Errorf("location %s has no IPv4 addresses available", l.Code)
		}
		return nil
	}
	return fmt.Errorf("location %q not found among %d locations", st.cfg.location, len(locs.Return))
}

// resolveImage finds a high-performance image code for the location.
//
// Codes carry a build date and change when images are rebuilt, so they
// are looked up rather than pinned. The catalogue is per-location and
// only reachable with the type and location filters together.
func resolveImage(ctx context.Context, c clients, st *state) error {
	time.Sleep(throttle)
	imgs, err := c.server.ListImages(ctx, server.ListImagesOptions{
		Type:     server.ImageTypeHPVMDistro,
		Location: st.cfg.location,
	})
	if err != nil {
		return fmt.Errorf("ListImages: %w", err)
	}
	if len(imgs.Return) == 0 {
		return fmt.Errorf("ListImages: no %s images at %s", server.ImageTypeHPVMDistro, st.cfg.location)
	}
	for _, im := range imgs.Return {
		// Rows with an empty code exist in the catalogue; skip them.
		if im.Code == "" {
			continue
		}
		if im.Distro == st.cfg.distro {
			st.cfg.image = im.Code
			log.Printf("✓ resolved %s at %s -> %s (%s)", st.cfg.distro, st.cfg.location, im.Code, im.Name)
			return nil
		}
	}
	return fmt.Errorf("ListImages: no image for distro %q at %s among %d candidate(s); set SH_DISTRO or SH_IMAGE",
		st.cfg.distro, st.cfg.location, len(imgs.Return))
}

// stepPreflight checks the product is offered at the location and has
// capacity.
//
// CanProvision is more useful than it first appears. It distinguishes
// three outcomes, and the difference matters because the recovery
// differs:
//
//	Successful                 offered here, and capacity is free
//	"No available nodes found" offered here, but currently full —
//	                           retry later, or pick another location
//	"Products not found"       not offered here at all — the code is
//	                           wrong for this location
//
// So it is a genuine existence-and-capacity check for a
// product/location pair. It is not how you find out what the codes
// are — [server.Client.ListProducts] does that, and stepDiscover calls
// it first. The two answer different questions: ListProducts says what
// exists and what it is made of, CanProvision says whether this one
// can be built here right now.
//
// What it does NOT validate is the image. It returns success for image
// codes that provision then rejects as unknown, so it cannot confirm a
// whole request is well-formed.
func stepPreflight(ctx context.Context, c clients, st *state) error {
	time.Sleep(throttle)
	resp, err := c.server.CanProvision(ctx, server.CanProvisionOptions{
		Product:  st.cfg.product,
		Location: st.cfg.location,
		Distro:   st.cfg.distro,
	})
	if err != nil {
		return fmt.Errorf("CanProvision: %w", err)
	}
	if !resp.Status {
		return fmt.Errorf("CanProvision: %s at %s unavailable: %s", st.cfg.product, st.cfg.location, resp.Msg)
	}
	log.Printf("✓ %s is offered at %s and has capacity", st.cfg.product, st.cfg.location)
	log.Printf("  note: this validates the product and capacity, not the image")
	return nil
}

// describeProduct reads the configured product out of the location's
// catalogue and reports what it is made of.
//
// This is the endpoint that removes the need to hardcode a product
// code, so an example that hardcodes one and never calls it would be
// teaching the opposite of what it documents. It also produces a
// better failure than CanProvision's "Products not found": the codes
// that ARE offered here.
//
// The partitions are the practical payoff. They name the disk labels
// ("scsi0" on high performance, "xvda1" on Xen) that
// server.UpgradeComponents requires, so a caller can know them before
// a server exists rather than only by reading Get afterwards.
func describeProduct(ctx context.Context, c clients, st *state) error {
	time.Sleep(throttle)
	prods, err := c.server.ListProducts(ctx, server.ListProductsOptions{
		Location: st.cfg.location,
	})
	if err != nil {
		return fmt.Errorf("ListProducts: %w", err)
	}

	codes := make([]string, 0, len(prods.Return))
	var found *server.Product
	for i, p := range prods.Return {
		codes = append(codes, p.Code)
		if p.Code == st.cfg.product {
			found = &prods.Return[i]
		}
	}
	if len(codes) == 0 {
		return fmt.Errorf("ListProducts: %s offers no products at all", st.cfg.location)
	}
	if found == nil {
		return fmt.Errorf("product %s is not offered at %s; that location offers %d code(s): %s",
			st.cfg.product, st.cfg.location, len(codes), strings.Join(codes, ", "))
	}

	log.Printf("✓ %s at %s: %d core(s), %.1fGB RAM, %dGB disk",
		found.Code, st.cfg.location,
		found.Attributes.Cores, found.Attributes.RAM, found.Attributes.Disk)

	// A loop over an empty list would log nothing and look like a pass,
	// so say which case this is.
	if len(found.Attributes.Partitions) == 0 {
		log.Printf("  no partitions reported; UpgradeComponents disk labels must come from server.Get")
		return nil
	}
	labels := make([]string, 0, len(found.Attributes.Partitions))
	for _, part := range found.Attributes.Partitions {
		labels = append(labels, part.Name)
	}
	log.Printf("✓ disk labels for UpgradeComponents, known before the server exists: %s",
		strings.Join(labels, ", "))
	return nil
}
