package main

import (
	"context"
	"fmt"
	"log"
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
// Standard-performance images are in the default catalogue instead, so
// when SH_PRODUCT names a standard product you must supply SH_IMAGE
// too — the two catalogues do not share codes, and a code from one is
// rejected by the other.
func stepDiscover(ctx context.Context, c clients, st *state) error {
	if err := checkLocation(ctx, c, st); err != nil {
		return err
	}
	if st.cfg.image != "" {
		log.Printf("✓ image pinned by SH_IMAGE: %s", st.cfg.image)
		return nil
	}
	return resolveImage(ctx, c, st)
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
// product/location pair, and the cheapest way to validate a product
// code without a products endpoint to list them.
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
