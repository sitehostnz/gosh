package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sitehostnz/gosh/pkg/api/server"
)

// stepProvision creates the pair and waits for both builds.
//
// Two things here catch people out.
//
// The label is not the name. The platform derives a name from the
// label, truncating it and appending a digit if it collides with an
// existing server — so provisioning "gosh-journey-a" and
// "gosh-journey-b" can yield "gosh-journey" and "gosh-journey1". Every
// later call identifies the server by name, so the returned Name is
// recorded here and the label is never used again.
//
// The SSH key is passed as public key content, not as the id returned
// when it was registered. Params.SSHKeys maps to params[ssh_keys][],
// which the API parses as key material.
func stepProvision(ctx context.Context, c clients, st *state) error {
	if st.cfg.image == "" {
		return fmt.Errorf("no image resolved: run step 20 (discover) first, or set SH_IMAGE")
	}
	if st.publicKey == "" {
		log.Printf("  no SSH key in state — provisioning without one; steps 50 and 80 will not be able to log in")
	}

	for _, label := range []string{st.cfg.labelA, st.cfg.labelB} {
		if err := createOne(ctx, c, st, label); err != nil {
			return err
		}
	}
	if st.nameA == "" || st.nameB == "" {
		return fmt.Errorf("expected two servers, got %v", st.created)
	}
	return readBack(ctx, c, st)
}

// createOne provisions a single server and records the name the
// platform assigned.
//
// The label is not the name: the platform truncates it and appends a
// digit on collision, so "gosh-journey-a" and "gosh-journey-b" can
// become "gosh-journey" and "gosh-journey1". Every later call uses the
// returned name.
//
// The SSH key goes in as public key content, not as the id returned
// when it was registered — Params.SSHKeys maps to params[ssh_keys][],
// which the API parses as key material.
func createOne(ctx context.Context, c clients, st *state, label string) error {
	time.Sleep(throttle)
	req := server.CreateRequest{
		Label:       label,
		Location:    st.cfg.location,
		ProductCode: st.cfg.product,
		Image:       st.cfg.image,
	}
	if st.publicKey != "" {
		req.Params.SSHKeys = []string{st.publicKey}
	}
	resp, err := c.server.Create(ctx, req)
	if err != nil {
		return fmt.Errorf("Create(%s): %w", label, err)
	}
	if resp.Return.Name == "" {
		return fmt.Errorf("Create(%s): API returned no server name", label)
	}
	name := resp.Return.Name
	// Record before waiting: if the build fails, step 90 still has
	// something to delete.
	st.created = append(st.created, name)
	if st.nameA == "" {
		st.nameA = name
	} else {
		st.nameB = name
	}
	log.Printf("✓ created label=%s -> name=%s ips=%v", label, name, resp.Return.Ips)

	if err := waitJob(ctx, c.job, resp.Return.Job); err != nil {
		return fmt.Errorf("Create(%s): %w", label, err)
	}
	return nil
}

// readBack records the addresses and product family the later steps
// branch on.
func readBack(ctx context.Context, c clients, st *state) error {
	time.Sleep(throttle)
	get, err := c.server.Get(ctx, server.GetRequest{ServerName: st.nameA})
	if err != nil {
		return fmt.Errorf("Get(%s): %w", st.nameA, err)
	}
	st.productType = get.Server.ProductType
	if st.productType == "" {
		return fmt.Errorf("Get(%s): server reports no product type", st.nameA)
	}

	if st.ipA, err = primaryIPv4(ctx, c.server, st.nameA); err != nil {
		return err
	}
	time.Sleep(throttle)
	if st.ipB, err = primaryIPv4(ctx, c.server, st.nameB); err != nil {
		return err
	}
	if st.ipA.IPAddr == st.ipB.IPAddr {
		return fmt.Errorf("both servers report the same primary %s", st.ipA.IPAddr)
	}

	// The account to log in as is not reported by the API and depends on
	// both the product family and the distro; resolve it once here so
	// steps 50 and 80 do not have to guess.
	if user, ok := server.LoginUserFor(st.productType, get.Server.Distro); ok {
		journeyLoginUser = user
		log.Printf("✓ platform: %s (%s), distro %s, login as %q",
			st.productType, get.Server.ProductCode, get.Server.Distro, user)
	} else {
		log.Printf("✓ platform: %s (%s), distro %s", st.productType, get.Server.ProductCode, get.Server.Distro)
		log.Printf("  no known login account for this product/distro pair; set SH_SSH_USER before steps 50 and 80")
	}
	log.Printf("✓ %s: %s network=%s gw=%s mac=%s (%s)", st.nameA, st.ipA.IPAddr, st.ipA.NetworkID, st.ipA.Gateway, st.ipA.MacAddr, subnet(st.ipA.IPAddr))
	log.Printf("✓ %s: %s network=%s gw=%s mac=%s (%s)", st.nameB, st.ipB.IPAddr, st.ipB.NetworkID, st.ipB.Gateway, st.ipB.MacAddr, subnet(st.ipB.IPAddr))
	if st.ipA.NetworkID == st.ipB.NetworkID {
		log.Printf("  both landed in the same network; the swap is unconstrained")
	} else {
		log.Printf("  different networks — the case that constrains the swap")
	}
	return nil
}
