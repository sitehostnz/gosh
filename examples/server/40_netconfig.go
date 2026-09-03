package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/sitehostnz/gosh/pkg/api/server"
)

// stepNetConfig fetches the guest network configuration for each
// server.
//
// This is the half of an address change the API cannot do for you.
// Reassigning an address does not touch the guest, and each address
// carries its own netmask and gateway, so until these files are written
// inside the server and it is rebooted, a swapped server does not come
// back. generate_network_config returns the files, keyed by path;
// getting them onto the guest is step 50's job.
func stepNetConfig(ctx context.Context, c clients, st *state) error {
	if err := st.resolveServers(); err != nil {
		return err
	}
	for _, name := range []string{st.nameA, st.nameB} {
		if name == "" {
			continue
		}
		time.Sleep(throttle)
		resp, err := c.server.GenerateNetworkConfig(ctx, server.GenerateNetworkConfigOptions{Name: name})
		if err != nil {
			return fmt.Errorf("GenerateNetworkConfig(%s): %w", name, err)
		}
		if len(resp.Return) == 0 {
			return fmt.Errorf("GenerateNetworkConfig(%s): no files returned; a swapped server would have nothing to write", name)
		}
		paths := make([]string, 0, len(resp.Return))
		for path, body := range resp.Return {
			if len(body) == 0 {
				return fmt.Errorf("GenerateNetworkConfig(%s): %s is empty", name, path)
			}
			paths = append(paths, path)
		}
		log.Printf("✓ %s: %d config file(s): %s", name, len(paths), strings.Join(paths, ", "))

		if err := compareWithGuest(st, name, resp.Return); err != nil {
			return err
		}
	}
	return nil
}

// compareWithGuest checks the platform's idea of a server's network
// configuration against the file actually on its disk.
//
// # Why this is worth doing
//
// Everything above asks the API to describe a server and then checks
// the API's answer, which establishes that the endpoint returns
// something well-formed. It does not establish that what it returned
// resembles the machine.
//
// That gap matters here more than most, because step 50 writes these
// files into the guests and step 70 reboots onto them. If the platform
// renders a path the guest does not use — a different netplan file, a
// different renderer, a distro that moved things between releases —
// step 50 would write a file nobody reads, the reboot would come up on
// the old addressing, and the failure would surface as an unreachable
// server with no clue why.
//
// So: read the same path out of the guest and compare. It needs a key,
// and skips with a reason when there is not one, because a check that
// quietly does nothing is worse than one that says it cannot run.
func compareWithGuest(st *state, name string, files map[string]string) error {
	if len(st.privateKey) == 0 && os.Getenv("SH_SSH_KEY_FILE") == "" {
		log.Printf("  no SSH key available; not comparing against the guest's own file")
		return nil
	}
	if len(st.privateKey) > 0 {
		journeyKey = st.privateKey
	}

	addr, err := addressOf(st, name)
	if err != nil {
		log.Printf("  no address known for %s; skipping the guest comparison", name)
		return nil //nolint:nilerr // absence of an address is not this step's failure
	}

	// Reachability first, so an unreachable guest is reported as that
	// rather than as a missing file.
	if !tcpReachable(addr, "22") {
		log.Printf("  %s is not reachable on %s yet; skipping the guest comparison", name, addr)
		return nil
	}

	for path := range files {
		// exists() reads test(1)'s exit status. Parsing stdout for a
		// word was tried and produced a false positive here: an empty
		// answer is a third state, and the parse read it as "absent".
		// That mistake had already been made and fixed once in the
		// snapshot step; writing this from memory rather than reusing
		// the helper reproduced it.
		if !exists(addr, path) {
			return fmt.Errorf("the platform renders %s for %s, but that path does not exist on the guest; step 50 would write a file nothing reads and the reboot in step 70 would come up on the old addressing", path, name)
		}
		log.Printf("  ✓ %s exists on %s, so step 50 will write a file the guest actually uses", path, name)
	}
	return nil
}

// addressOf returns the address this journey knows for a server.
func addressOf(st *state, name string) (string, error) {
	switch name {
	case st.nameA:
		if st.ipA.IPAddr != "" {
			return st.ipA.IPAddr, nil
		}
	case st.nameB:
		if st.ipB.IPAddr != "" {
			return st.ipB.IPAddr, nil
		}
	}
	return "", fmt.Errorf("no address recorded for %s", name)
}
