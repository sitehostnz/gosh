package main

import (
	"context"
	"fmt"
	"log"
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
// getting them onto the guest is step 60's job.
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
	}
	return nil
}
