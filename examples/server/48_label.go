package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sitehostnz/gosh/pkg/api/server"
)

// stepLabel changes a server's label and puts it back.
//
// # Name and label are different things
//
// The label is the display name a caller chooses. The name is what the
// platform assigns, constrained to 2-13 characters, and it is what
// every other endpoint takes. Provisioning derives the name from the
// label by truncating it, which is why two servers labelled
// "gosh-journey-a" and "gosh-journey-b" end up named "gosh-journey"
// and "gosh-journe1".
//
// Update changes only the label. There is no endpoint that renames a
// server, so a name is permanent for the life of the server — worth
// knowing before choosing a labelling scheme that relies on names
// matching.
//
// It refuses to run against a server this process did not create, for
// the same reason the delete step does.
func stepLabel(ctx context.Context, c clients, st *state) error {
	if len(st.created) == 0 {
		return fmt.Errorf("the label step changes a server's label; run the journey so it acts on a server this process created")
	}
	name := st.created[0]

	time.Sleep(throttle)
	before, err := c.server.Get(ctx, server.GetRequest{ServerName: name})
	if err != nil {
		return fmt.Errorf("server.Get: %w", err)
	}
	original := before.Server.Label
	log.Printf("  %s is labelled %q", name, original)

	changed := original + "-relabelled"
	time.Sleep(throttle)
	if _, err := c.server.Update(ctx, server.UpdateRequest{Name: name, Label: changed}); err != nil {
		return fmt.Errorf("server.Update: %w", err)
	}

	after, err := readLabel(ctx, c, name)
	if err != nil {
		return err
	}
	if after != changed {
		return fmt.Errorf("label is %q after the update, want %q; the change reported success without taking effect", after, changed)
	}
	log.Printf("✓ label changed to %q", changed)

	if err := assertNameUnchanged(ctx, c, name); err != nil {
		return err
	}

	time.Sleep(throttle)
	if _, err := c.server.Update(ctx, server.UpdateRequest{Name: name, Label: original}); err != nil {
		return fmt.Errorf("server.Update restoring the label: %w", err)
	}
	restored, err := readLabel(ctx, c, name)
	if err != nil {
		return err
	}
	if restored != original {
		return fmt.Errorf("label is %q after restoring, want %q", restored, original)
	}
	log.Printf("✓ restored to %q", original)
	return nil
}

// readLabel reads a server's current label.
func readLabel(ctx context.Context, c clients, name string) (string, error) {
	time.Sleep(throttle)
	got, err := c.server.Get(ctx, server.GetRequest{ServerName: name})
	if err != nil {
		return "", fmt.Errorf("server.Get: %w", err)
	}
	return got.Server.Label, nil
}

// assertNameUnchanged checks relabelling did not rename the server.
//
// This is the assumption a caller is most likely to get wrong, and the
// check that would catch it: if the name had moved with the label, the
// Get below would fail to find the server at all.
func assertNameUnchanged(ctx context.Context, c clients, name string) error {
	time.Sleep(throttle)
	got, err := c.server.Get(ctx, server.GetRequest{ServerName: name})
	if err != nil {
		return fmt.Errorf("server.Get after relabelling — the name may have changed with the label: %w", err)
	}
	if got.Server.Name != name {
		return fmt.Errorf("server name is now %q, was %q; relabelling is not supposed to rename", got.Server.Name, name)
	}
	log.Printf("✓ the name is unchanged; only the label moved")
	return nil
}
