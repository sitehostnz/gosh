package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/sitehostnz/gosh/pkg/api/cloud/db"
	dbuser "github.com/sitehostnz/gosh/pkg/api/cloud/db/user"
	sshuser "github.com/sitehostnz/gosh/pkg/api/cloud/ssh/user"
	"github.com/sitehostnz/gosh/pkg/api/cloud/stack"
	"github.com/sitehostnz/gosh/pkg/api/server"
	sshkey "github.com/sitehostnz/gosh/pkg/api/ssh/key"
)

// stepDelete tears down everything this process created.
//
// # What it will and will not delete
//
// Only servers in state.created — the ones this process provisioned.
// A server named by SH_SERVER is never deleted, however the journey was
// invoked. Naming an existing server so a read step has something to
// look at must not be the same act as marking it for destruction.
//
// Deleting the container server removes its stacks, databases and SSH
// users with it, so the individual deletes below are not strictly
// required to reclaim anything. They are done anyway, in the reverse of
// the order they were created, because they are the only thing that
// exercises those delete paths at all — and because a delete that only
// ever runs implicitly is a delete nobody has tested.
//
// # Each one is waited for, and the order is load-bearing
//
// Every delete here returns a job, and the platform holds short locks
// between them. Not waiting produced two distinct failures in one run:
// removing the account SSH key while the SSH user's delete was still
// running was refused with "there is a job already running on this
// user", and deleting the key after the server was gone failed with an
// unrecoverable HTTP 500. The key therefore goes after the SSH user has
// finished and before the server is touched.
//
// A failure in any one of them is reported and does not stop the
// others: the server delete is what actually reclaims the resource, and
// skipping it because a stack delete failed would leave the thing
// running.
func stepDelete(ctx context.Context, c clients, st *state) error {
	targets := deletable(st)
	if len(targets) == 0 {
		log.Printf("  nothing to delete: this process provisioned no container server")
		return nil
	}

	var failures []error
	for _, name := range targets {
		for _, part := range teardown(st) {
			if part.skip {
				continue
			}
			if err := retryDelete(ctx, c, name, part); err != nil {
				log.Printf("  %s delete failed: %v", part.what, err)
				failures = append(failures, fmt.Errorf("%s: %w", part.what, err))
				continue
			}
			log.Printf("✓ deleted %s", part.what)
		}
		// Before the server, deliberately. See deleteAccountKey.
		if err := deleteAccountKey(ctx, c, st); err != nil {
			log.Printf("  %v", err)
			failures = append(failures, err)
		}
		if err := deleteServer(ctx, c, st, name); err != nil {
			failures = append(failures, err)
		}
	}

	if len(failures) > 0 {
		return fmt.Errorf("teardown had %d failure(s): %v", len(failures), failures)
	}
	return nil
}

// teardownTimeout bounds how long one delete is retried through the
// platform's transient refusals.
const teardownTimeout = 4 * time.Minute

// transientRefusals are the messages that mean "not yet" rather than
// "no".
//
// Each was met in a single teardown, and none of them is a failure of
// the delete being attempted:
//
//   - a job still running on the stack, queued by the database work
//     that preceded it;
//   - an SSH user with sessions still open — this journey's own
//     verification sessions, closed client-side but not yet reaped by
//     the host;
//   - a job still running on the SSH user, typically its own delete
//     from a previous attempt.
//
// Retrying is correct for all three. What must not be retried is
// anything else, so the match is on these phrases rather than on
// "the delete failed".
var transientRefusals = []string{
	"job already running",
	"cannot be deleted right now",
	"logged out",
	"is currently locked",
}

// isTransientRefusal reports whether an error is worth waiting out.
//
// Matching on message text, deliberately. These are the API's own
// rejections arriving as status:false with free-text messages, and
// there is no typed error to match on — unlike a transport failure,
// where errors.As on *url.Error and net.Error separates the cases by
// construction and substring matching has been got wrong in both
// directions. The distinction is worth keeping straight: classify
// transport errors structurally, and API messages by their text
// because that is all the API gives you.
func isTransientRefusal(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	for _, phrase := range transientRefusals {
		if strings.Contains(msg, phrase) {
			return true
		}
	}
	return false
}

// retryDelete runs one teardown part, waiting out transient refusals.
//
// The first version of this ran each delete once. Three of the five
// parts failed in one run — on a stack with a queued job, on an SSH
// user with sessions still open, and on a database user whose row had
// gone because the database was deleted first — and every one of them
// would have succeeded moments later. A teardown that gives up on the
// first "not yet" leaves real resources behind, which is worse than
// slow.
func retryDelete(ctx context.Context, c clients, name string, part teardownPart) error {
	deadline := time.Now().Add(teardownTimeout)
	for {
		time.Sleep(throttle)
		err := part.run(ctx, c, name)
		if err == nil {
			return nil
		}
		if !isTransientRefusal(err) {
			return err
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("still refused after %s: %w", teardownTimeout, err)
		}
		log.Printf("  %s: not yet (%v); waiting", part.what, err)
		time.Sleep(15 * time.Second)
	}
}

// teardownPart is one delete, with what to call it and whether there is
// anything to do.
type teardownPart struct {
	what string
	skip bool
	run  func(context.Context, clients, string) error
}

// teardown lists the deletes in reverse creation order.
//
// The database user comes before the database, and that way round
// matters: deleting the database first takes the user's grant row with
// it, and the user delete then reports "the database user couldn't be
// found or you don't have permission to delete it" — which reads like a
// permissions problem rather than an ordering one.
func teardown(st *state) []teardownPart {
	return []teardownPart{
		{
			what: fmt.Sprintf("database user %q", st.dbUser),
			skip: st.dbUser == "" || st.mysqlHost == "",
			run: func(ctx context.Context, c clients, name string) error {
				resp, err := c.dbUser.Delete(ctx, dbuser.DeleteRequest{
					ServerName: name, MySQLHost: st.mysqlHost, Username: st.dbUser,
				})
				if err != nil {
					return err
				}
				return waitJob(ctx, c.job, resp.Return.Job)
			},
		},
		{
			what: fmt.Sprintf("database %q", st.database),
			// Both, not just the name: the delete needs a host, and
			// attempting it without one returned "The mysql_host
			// parameter is missing" for a database that had never been
			// created.
			skip: st.database == "" || st.mysqlHost == "",
			run: func(ctx context.Context, c clients, name string) error {
				resp, err := c.db.Delete(ctx, db.DeleteRequest{
					ServerName: name, MySQLHost: st.mysqlHost, Database: st.database,
				})
				if err != nil {
					return err
				}
				return waitJob(ctx, c.job, resp.Return.Job)
			},
		},
		{
			what: fmt.Sprintf("stack %q", st.stackName),
			skip: st.stackName == "",
			run: func(ctx context.Context, c clients, name string) error {
				resp, err := c.stack.Delete(ctx, stack.DeleteRequest{ServerName: name, Name: st.stackName})
				if err != nil {
					return err
				}
				return waitJob(ctx, c.job, resp.Return.Job)
			},
		},
		{
			what: fmt.Sprintf("ssh user %q", st.sshUser),
			skip: st.sshUser == "",
			run: func(ctx context.Context, c clients, name string) error {
				resp, err := c.sshers.Delete(ctx, sshuser.DeleteRequest{ServerName: name, Username: st.sshUser})
				// The account name is recorded before it is created, so
				// that a partially-succeeded Add still gets cleaned up.
				// The cost is that an Add which failed outright leaves a
				// name here that never existed, and deleting it reports
				// "the SSH user couldn't be found". That is the state we
				// wanted, so it is not a teardown failure.
				if err != nil && strings.Contains(err.Error(), "couldn't be found") {
					log.Printf("  ssh user %q did not exist, which is the desired end state", st.sshUser)
					return nil
				}
				if err != nil {
					return err
				}
				// Waited on deliberately. Every delete here returns a
				// job, and this one has to finish before the account
				// SSH key can be removed: removing the key while the
				// user's delete was still running failed with "there is
				// a job already running on this user", and the key was
				// then left on the account.
				return waitJob(ctx, c.job, resp.Return.Job)
			},
		},
	}
}

// deleteAccountKey removes the SSH key this run registered.
//
// The key is account-level, not server-level: deleting the container
// server leaves it behind. Nothing else in the journey cleans it up, so
// a run that skipped this would leave one dead key on the account per
// invocation.
//
// # It has to happen before the server is deleted
//
// verified live against API 1.5.
//
// Removing a key enumerates every container SSH user that references
// it and updates each one. If a referencing user belongs to a server
// that has since been deleted, that lookup fails and the call returns
// an empty HTTP 500 — and the key cannot be removed afterwards either,
// by retrying or later. Keys removed while their server was still up
// came off without trouble.
//
// So this runs first. The ordering is not tidiness: it is the
// difference between the journey cleaning up after itself and leaking a
// key per run with no way to remove it.
func deleteAccountKey(ctx context.Context, c clients, st *state) error {
	if st.sshKeyID == "" {
		return nil
	}
	time.Sleep(throttle)
	if _, err := c.sshkeys.Delete(ctx, sshkey.DeleteRequest{ID: st.sshKeyID}); err != nil {
		return fmt.Errorf("ssh/key.Delete(%s): %w", st.sshKeyID, err)
	}
	log.Printf("✓ deleted account ssh key %s", st.sshKeyID)
	return nil
}

// deleteServer removes the container server and proves it stopped
// answering.
//
// The delete returning is the control plane accepting the request. The
// out-of-band half is tcp/22 ceasing to answer: while the address still
// completes a handshake, something is still running on it, whatever the
// API says. That check is bounded, and a timeout is reported rather
// than raised — the delete is asynchronous and the address can be
// reclaimed on its own schedule, so a still-open port here means "not
// yet", not "the delete failed".
func deleteServer(ctx context.Context, c clients, st *state, name string) error {
	time.Sleep(throttle)
	resp, err := c.server.Delete(ctx, server.DeleteRequest{Name: name})
	if err != nil {
		// A build that fails is cleaned up by the platform, which
		// deletes the half-made server itself. The name is recorded
		// before the wait — deliberately, so a server that came up and
		// then failed still gets torn down — so this path is reached
		// with nothing left to delete. "Not Found" is the end state
		// this function exists to reach, not a teardown failure.
		if strings.Contains(err.Error(), "Not Found") {
			log.Printf("  %s no longer exists, which is the desired end state", name)
			return nil
		}
		return fmt.Errorf("server.Delete(%s): %w", name, err)
	}
	log.Printf("✓ requested delete of %s (job %d)", name, resp.Return.ID)

	if err := waitJob(ctx, c.job, resp.Return.Job); err != nil {
		return fmt.Errorf("server.Delete(%s): %w", name, err)
	}
	log.Printf("✓ delete job completed for %s", name)

	if st.ip == "" {
		return nil
	}
	gone, took := waitReachability(st.ip, "22", false)
	if gone {
		log.Printf("✓ out of band: %s stopped answering on tcp/22 after %s", st.ip, took.Round(time.Second))
	} else {
		log.Printf("  %s still answers on tcp/22 after %s; the address may not be reclaimed yet", st.ip, took.Round(time.Second))
	}
	return nil
}
