package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sitehostnz/gosh/pkg/api/server"
)

// stepReboot restarts both servers through the API so they pick up the
// configuration staged in step 50.
//
// This is the step that makes the whole cutover safe. Rebooting through
// change_state needs no access to the guest, so it works fine on a
// server whose address has just moved and which nothing can reach. That
// removes any need to time a reboot against the swap, or to arm a
// delayed one before it.
//
// The sequence that matters: stage the config (50), move the address
// (60), then reboot (70). Each server comes up with the addressing it
// now actually owns.
func stepReboot(ctx context.Context, c clients, st *state) error {
	if err := st.resolveServers(); err != nil {
		return err
	}
	for _, name := range []string{st.nameA, st.nameB} {
		if name == "" {
			continue
		}
		time.Sleep(throttle)
		resp, err := c.server.ChangeState(ctx, server.ChangeStateOptions{
			Name:  name,
			State: "reboot",
		})
		if err != nil {
			return fmt.Errorf("ChangeState(%s, reboot): %w", name, err)
		}
		if err := waitJob(ctx, c.job, resp.Return.Job); err != nil {
			return fmt.Errorf("ChangeState(%s, reboot): %w", name, err)
		}
		log.Printf("✓ %s: reboot requested and job completed", name)
	}
	log.Printf("  both servers are restarting onto their new addressing")
	return nil
}
