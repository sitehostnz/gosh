package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"strconv"
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
//   - The image code is not checked when the request is made, but it
//     is not ignored either. An unknown code returns a job id as
//     normal, then the build sits in "Configuring server" for twenty
//     minutes, fails, and the platform deletes the half-built server.
//     A finished container reports image as null, which is what makes
//     "it does not take an image" a tempting and expensive conclusion.
//     This step therefore checks the code against server.ListImages
//     before anything is provisioned.
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

	if err := checkImage(ctx, c, st); err != nil {
		return err
	}
	return pickProduct(ctx, c, st)
}

// checkImage fails fast if the configured image code does not exist.
//
// This is the cheapest check in the journey and it guards the most
// expensive failure: an unknown code is accepted by the provision call
// and only fails twenty minutes later, by which time the server it was
// building has been deleted and the job says nothing more useful than
// Failed.
func checkImage(ctx context.Context, c clients, st *state) error {
	if st.imageChecked {
		return nil
	}
	time.Sleep(throttle)
	images, err := c.server.ListImages(ctx, server.ListImagesOptions{})
	if err != nil {
		return fmt.Errorf("ListImages: %w", err)
	}

	codes := make([]string, 0, len(images.Return))
	for _, img := range images.Return {
		if img.Code == st.cfg.image {
			log.Printf("✓ image %s exists (%s, %s)", img.Code, img.Name, img.Type)
			st.imageChecked = true
			return nil
		}
		codes = append(codes, img.Code)
	}
	return fmt.Errorf("image %q is not in the catalogue, and provisioning with an unknown code fails silently twenty minutes later; available: %v",
		st.cfg.image, codes)
}

// pickProduct resolves the product code the journey will provision.
//
// SH_PRODUCT pins a code. With nothing pinned the smallest one offered
// at the location is chosen, by core count and then by price, rather
// than a code written into the source. There are 21 CLDCON codes at
// AKLNCT and the set differs by location, so a hardcoded default is a
// default that is wrong somewhere.
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
		switch {
		case st.cfg.product != "":
			if p.Code == st.cfg.product {
				found = &prods.Return[i]
			}
		case found == nil, smaller(prods.Return[i], *found):
			found = &prods.Return[i]
		}
	}
	if found == nil {
		return fmt.Errorf("product %s not offered at %s; available: %v", st.cfg.product, st.cfg.location, codes)
	}
	st.cfg.product = found.Code

	log.Printf("✓ %s: %d core(s), %.0fGB RAM, %s/month",
		found.Code, found.Attributes.Cores, found.Attributes.RAM, found.Price.String())

	// Containers report no disk attributes, unlike virtual servers.
	if found.Attributes.Disk == 0 && len(found.Attributes.Partitions) == 0 {
		log.Printf("  no disk or partitions reported — containers are sized by cores and RAM")
	}
	return nil
}

// smaller reports whether a is a smaller product than b, by cores then
// price. Both are compared so that two codes with equal cores — and
// AKLNCT offers several — resolve deterministically instead of by map
// order.
func smaller(a, b server.Product) bool {
	if a.Attributes.Cores != b.Attributes.Cores {
		return a.Attributes.Cores < b.Attributes.Cores
	}
	return priceOf(a) < priceOf(b)
}

// priceOf parses a product's price for comparison.
//
// Price arrives as shtypes.MaybeString because the API has sent it both
// quoted and bare. An unparseable price sorts last rather than erroring:
// this is choosing between products, not billing.
func priceOf(p server.Product) float64 {
	v, err := strconv.ParseFloat(p.Price.String(), 64)
	if err != nil {
		return math.MaxFloat64
	}
	return v
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
