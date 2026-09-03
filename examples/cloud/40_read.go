package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/sitehostnz/gosh/pkg/api/cloud/db"
	dbuser "github.com/sitehostnz/gosh/pkg/api/cloud/db/user"
	cloudserver "github.com/sitehostnz/gosh/pkg/api/cloud/server"
	sshuser "github.com/sitehostnz/gosh/pkg/api/cloud/ssh/user"
	"github.com/sitehostnz/gosh/pkg/api/cloud/stack"
	stackimage "github.com/sitehostnz/gosh/pkg/api/cloud/stack/image"
)

// stepRead walks every read path on one existing cloud server.
//
// # Why this is a step and not a test
//
// It is the half of the recording corpus the probes cannot supply. The
// probes establish what the API rejects; this establishes the shape of
// what it returns when it agrees — which is the half that catches a
// field typed as the wrong thing. Both are needed: a decode bug hides
// in a success, a request bug hides in a rejection.
//
// It reads only. Nothing here creates, changes or deletes anything, so
// it needs no opt-in. It does need a server to look at: set SH_SERVER
// to the name of a cloud container you already have, or the step picks
// the first one on the account.
//
// # Logging discipline
//
// Counts, ids and shapes only — never a database name, username, key or
// path. The recordings under SH_RECORD_DIR do hold real values, because
// they are the raw exchange; anything derived from them for a fixture
// has to be scrubbed first.
func stepRead(ctx context.Context, c clients, st *state) error {
	name, err := resolveServer(ctx, c, st)
	if err != nil {
		return err
	}
	log.Printf("  reading %s", name)

	reads := []struct {
		what string
		run  func(context.Context, clients, string) error
	}{
		{"update window", readUpdateWindow},
		{"stacks", readStacks},
		{"databases", readDatabases},
		{"database users", readDatabaseUsers},
		{"SSH users", readSSHUsers},
		{"stack images", readStackImages},
	}
	for _, r := range reads {
		if err := r.run(ctx, c, name); err != nil {
			return fmt.Errorf("%s: %w", r.what, err)
		}
	}
	return nil
}

// readUpdateWindow reads the maintenance window.
//
// The update window is a managed-service feature. On an unmanaged
// server the API answers "This server is not managed by SiteHost." — a
// property of the server, not a fault, so it is reported and the walk
// continues.
func readUpdateWindow(ctx context.Context, c clients, name string) error {
	time.Sleep(throttle)
	win, err := c.cloud.GetUpdateWindow(ctx, cloudserver.GetUpdateWindowRequest{ServerName: name})
	switch {
	case err != nil && strings.Contains(err.Error(), "not managed"):
		log.Printf("  no update window: this server is unmanaged")
	case err != nil:
		return err
	default:
		log.Printf("✓ update window: enabled=%t day=%d hour=%d",
			win.Return.Enabled, win.Return.DayOfWeek, win.Return.HourOfDay)
	}
	return nil
}

// readStacks lists the stacks and reads one back.
func readStacks(ctx context.Context, c clients, name string) error {
	time.Sleep(throttle)
	stacks, err := c.stack.List(ctx, stack.ListRequest{ServerName: name})
	if err != nil {
		return fmt.Errorf("stack.List: %w", err)
	}
	log.Printf("✓ %d stack(s)", len(stacks.Return.Stacks))

	// A loop over an empty list logs a tick it never earned, so say so.
	if len(stacks.Return.Stacks) == 0 {
		log.Printf("  no stacks; skipping stack.Get")
		return nil
	}

	time.Sleep(throttle)
	one := stacks.Return.Stacks[0]
	got, err := c.stack.Get(ctx, stack.GetRequest{ServerName: name, Name: one.Name})
	if err != nil {
		return fmt.Errorf("stack.Get %q: %w", one.Name, err)
	}
	if got.Stack.Name != one.Name {
		return fmt.Errorf("stack.Get returned %q, asked for %q", got.Stack.Name, one.Name)
	}
	log.Printf("✓ stack.Get agrees with the listing; compose file %d bytes", len(got.Stack.DockerFile))
	return nil
}

// readDatabases lists the databases and reads one back.
func readDatabases(ctx context.Context, c clients, name string) error {
	time.Sleep(throttle)
	dbs, err := c.db.List(ctx, db.ListOptions{ServerName: name})
	if err != nil {
		return fmt.Errorf("db.List: %w", err)
	}
	log.Printf("✓ %d database(s)", len(dbs.Return.Databases))

	if len(dbs.Return.Databases) == 0 {
		log.Printf("  no databases; skipping db.Get")
		return nil
	}

	time.Sleep(throttle)
	one := dbs.Return.Databases[0]
	got, err := c.db.Get(ctx, db.GetRequest{
		ServerName: name, MySQLHost: one.MySQLHost, Database: one.DBName,
	})
	if err != nil {
		return fmt.Errorf("db.Get: %w", err)
	}
	if got.Database.DBName != one.DBName {
		return fmt.Errorf("db.Get returned a different database than the one requested")
	}
	log.Printf("✓ db.Get agrees with the listing")
	return nil
}

// readDatabaseUsers lists the database users and checks the one
// invariant worth checking: passwords are never returned.
func readDatabaseUsers(ctx context.Context, c clients, name string) error {
	time.Sleep(throttle)
	users, err := c.dbUser.List(ctx, dbuser.ListOptions{ServerName: name})
	if err != nil {
		return fmt.Errorf("dbUser.List: %w", err)
	}
	log.Printf("✓ %d database user(s)", len(users.Return.Users))
	for _, u := range users.Return.Users {
		if u.Password != "" {
			return fmt.Errorf("a listing returned a non-empty password; passwords are supposed to be write-only")
		}
	}
	if len(users.Return.Users) > 0 {
		log.Printf("  every listed password is empty, as expected")
	}
	return nil
}

// readSSHUsers lists the SSH users and their keys.
func readSSHUsers(ctx context.Context, c clients, name string) error {
	time.Sleep(throttle)
	users, err := c.sshers.List(ctx, sshuser.ListOptions{ServerName: name})
	if err != nil {
		return fmt.Errorf("sshUser.List: %w", err)
	}
	var withKeys int
	for _, u := range users.Return.Users {
		if len(u.SSHKeys) > 0 {
			withKeys++
		}
	}
	log.Printf("✓ %d SSH user(s), %d with at least one key", len(users.Return.Users), withKeys)
	return nil
}

// readStackImages reads the catalogue and one image from it.
func readStackImages(ctx context.Context, c clients, _ string) error {
	time.Sleep(throttle)
	images, err := c.image.List(ctx)
	if err != nil {
		return fmt.Errorf("image.List: %w", err)
	}
	if len(images.Return) == 0 {
		return fmt.Errorf("image.List returned nothing; the catalogue is never empty")
	}
	log.Printf("✓ %d stack image(s) in the catalogue", len(images.Return))

	time.Sleep(throttle)
	code := images.Return[0].Code
	img, err := c.image.Get(ctx, stackimage.GetRequest{Code: code})
	if err != nil {
		return fmt.Errorf("image.Get %q: %w", code, err)
	}
	if img == nil || img.Code != code {
		return fmt.Errorf("image.Get returned a different image than the one requested")
	}
	log.Printf("✓ image.Get agrees with the catalogue")
	return nil
}

// resolveServer returns the server to read, from SH_SERVER, from state,
// or the first cloud server on the account.
func resolveServer(ctx context.Context, c clients, st *state) (string, error) {
	if st.name != "" {
		return st.name, nil
	}
	if n := envOr("SH_SERVER", ""); n != "" {
		st.name = n
		return n, nil
	}
	time.Sleep(throttle)
	list, err := c.cloud.List(ctx)
	if err != nil {
		return "", fmt.Errorf("cloud.List: %w", err)
	}
	if len(list.CloudServers) == 0 {
		return "", fmt.Errorf("no cloud servers on this account; set SH_SERVER or run the provision step")
	}
	st.name = list.CloudServers[0].Name
	log.Printf("  no SH_SERVER set; using the first cloud server on the account")
	return st.name, nil
}
