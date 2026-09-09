package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sitehostnz/gosh/pkg/api/server"
)

// stepProvision creates a Cloud Container and proves it came up.
//
// # A container server is provisioned through the server endpoints
//
// There is no cloud/create call. A Cloud Container is server.Create
// with a CLDCON product code, and from then on it is addressed by
// server name through both the server and the cloud packages. That is
// why this journey depends on the same discovery, job-polling and
// delete machinery as examples/server.
//
// # The label is not the name
//
// The platform derives the server name from the label, truncating it
// and appending a digit on collision, so provisioning
// "gosh-cloud-journey" twice yields "gosh-cloud-journey" and
// "gosh-cloud-journey1". Every later call in this journey uses the
// returned name.
//
// # No SSH key is sent here
//
// params[ssh_keys][] is accepted by the endpoint, and on a virtual
// server it installs the key for the login account. A Cloud Container
// has no such account: SSH access is granted per container SSH user
// through cloud/ssh/user.Add, so a key passed here reaches nothing.
// Step 30 generates the pair and registers it where it is actually
// used, which is also what lets steps 30 to 50 run against a server
// named by SH_SERVER.
//
// # What is verified out of band
//
// The job completing says the platform finished its work. It does not
// say the host answers. This step therefore opens a real socket to
// tcp/22 from wherever the example runs, and fails if it never comes
// up. That check is the reason the rest of the journey can trust its
// address.
//
// # A completed build job is not a server that is on
//
// Three things happen at different times, and code that assumes they
// are the same fails intermittently:
//
//   - the provision job reports Completed
//   - tcp/22 starts answering
//   - server.Get reports state "On"
//
// tcp/22 was observed answering while the state was still
// "Provisioning", and cloud/stack/add.json rejects a server that is not
// on — "Unable to add stack, the server must be on." So this step waits
// for the state as well as the port, because step 30 cannot proceed
// without it.
//
// tcp/80 is reported here rather than asserted, and the reason is
// worth stating because the obvious explanation is wrong.
//
// Port 80 was observed refused on this server and then, once the same
// server finished building, answering with no stacks deployed at all.
// The refusal tracked the build, not the absence of a website
// container: the host's reverse proxy listens from the start and
// answers 503 until something is routable. So a port-80 handshake says
// the proxy is up, and says nothing about whether any container
// exists — which makes it a poor assertion for this step and no
// substitute for step 30 actually fetching a page.
func stepProvision(ctx context.Context, c clients, st *state) error {
	if err := ensureProvisionInputs(ctx, c, st); err != nil {
		return err
	}
	// Announced before the call, not after. The label is randomly
	// generated and is the only handle on what this run is about to
	// create; logging it only on success means a create that reaches
	// the platform and loses its response leaves a container server
	// nobody can name, recoverable only by listing the account. The
	// derived name is "ch-" plus the label's first nine characters, so
	// this line is enough to find it.
	log.Printf("  creating a container server from label %q at %s (%s)",
		st.cfg.label, st.cfg.location, st.cfg.product)

	time.Sleep(throttle)
	req := server.CreateRequest{
		Label:       st.cfg.label,
		Location:    st.cfg.location,
		ProductCode: st.cfg.product,
		Image:       st.cfg.image,
	}
	resp, err := c.server.Create(ctx, req)
	if err != nil {
		return fmt.Errorf("Create(%s): %w", st.cfg.label, err)
	}
	if resp.Return.Name == "" {
		return fmt.Errorf("Create(%s): API returned no server name", st.cfg.label)
	}

	// Record before waiting. If the build fails or the wait times out,
	// step 90 still has something to delete.
	st.name = resp.Return.Name
	st.created = append(st.created, st.name)
	log.Printf("✓ created label=%s -> name=%s product=%s", st.cfg.label, st.name, st.cfg.product)

	if err := waitJob(ctx, c.job, resp.Return.Job); err != nil {
		return fmt.Errorf("Create(%s): %w", st.cfg.label, err)
	}
	log.Printf("✓ build job completed")

	return confirmUp(ctx, c, st)
}

// confirmUp reads the address back and proves the host answers.
//
// The two assertions are different in kind. server.Get reporting an
// address is the control plane describing its own record. A completed
// TCP handshake to that address is the host itself answering, from
// outside the platform. Only the second can fail in a way that tells
// you the provision did not really work.
func confirmUp(ctx context.Context, c clients, st *state) error {
	time.Sleep(throttle)
	ip, err := primaryIPv4(ctx, c.server, st.name)
	if err != nil {
		return err
	}
	st.ip = ip

	time.Sleep(throttle)
	get, err := c.server.Get(ctx, server.GetRequest{ServerName: st.name})
	if err != nil {
		return fmt.Errorf("Get(%s): %w", st.name, err)
	}
	if got := get.Server.ProductType; got != "CLDCON" {
		return fmt.Errorf("Get(%s): product type is %q, want CLDCON — this is not a container server", st.name, got)
	}
	log.Printf("✓ %s: %s product_type=%s distro=%s state=%s",
		st.name, st.ip, get.Server.ProductType, get.Server.Distro, get.Server.State)

	// The container took no image, and reports none: server/get.json
	// sends "image": null for a CLDCON. models.Server has no Image
	// field to receive it, so there is nothing to assert here — noted
	// because the absence is easy to mistake for a decode fault.
	log.Printf("  image sent as %q and unused; a container reports no image", st.cfg.image)

	if err := waitForOn(ctx, c, st); err != nil {
		return err
	}

	ok, took := waitReachability(st.ip, "22", true)
	if !ok {
		return fmt.Errorf("%s: tcp/22 never answered within %s — the job completed but the host is not up", st.ip, probeSettle)
	}
	log.Printf("✓ out of band: tcp/22 on %s answered after %s", st.ip, took.Round(time.Second))

	// Reported, not asserted: the proxy answers on 80 from the start,
	// with or without a container behind it. Step 30 is where fetching
	// a page becomes a real check.
	if tcpReachable(st.ip, "80") {
		log.Printf("  tcp/80 answers — that is the host's proxy, not a container")
	} else {
		log.Printf("  tcp/80 refuses, which on a finished container server is unexpected")
	}
	return nil
}

// ensureProvisionInputs resolves what a standalone run has not been
// told.
//
// SH_PRODUCT pins a code; without one the catalogue is consulted the
// same way step 10 does. The image check runs either way, because it is
// the guard that turns a typo from a twenty-minute silent failure into
// an immediate error, and a step that provisions must not be able to
// skip it by being run on its own.
func ensureProvisionInputs(ctx context.Context, c clients, st *state) error {
	if err := checkImage(ctx, c, st); err != nil {
		return err
	}
	if st.cfg.product != "" {
		return nil
	}
	return pickProduct(ctx, c, st)
}

// waitForOn polls until the server reports state "On".
//
// The build job completing is not sufficient: the state lags it, and
// the stack endpoints reject a server that is not on. Polling the state
// here means step 30 fails for real reasons rather than for being
// early.
func waitForOn(ctx context.Context, c clients, st *state) error {
	deadline := time.Now().Add(jobTimeout)
	var last string
	for {
		time.Sleep(throttle * 2)
		get, err := c.server.Get(ctx, server.GetRequest{ServerName: st.name})
		if err != nil {
			// A poll can fail on the per-second limit; that is not the
			// server failing to come up.
			if time.Now().After(deadline) {
				return fmt.Errorf("waiting for %s to be On: %w", st.name, err)
			}
			continue
		}
		if get.Server.State == "On" {
			log.Printf("✓ %s reports state On", st.name)
			return nil
		}
		if get.Server.State != last {
			log.Printf("  %s is %s", st.name, get.Server.State)
			last = get.Server.State
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("%s is still %s after %s", st.name, get.Server.State, jobTimeout)
		}
	}
}
