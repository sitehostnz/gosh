package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/sitehostnz/gosh/pkg/api/server"
	sshkey "github.com/sitehostnz/gosh/pkg/api/ssh/key"
)

// deleteRetries bounds how long to keep retrying a delete that is
// blocked by a build still in progress.
const deleteRetries = 20

// stepDelete removes everything this run created.
//
// Always last, and always attempted: the journey runs this even when an
// earlier step failed, because a server left behind keeps costing money
// and a half-finished run is exactly when one gets forgotten.
//
// Two behaviours make this less trivial than it looks:
//
//   - A server cannot be deleted while it is in a transitional state.
//     The API rejects it with "The specified server cannot be deleted
//     while in the '<state>' state" — 'Provisioning' after a build,
//     'Shutting Down' after a reboot — so this retries rather than
//     giving up on what would otherwise be a leak. The match is on the
//     shape of that message rather than a list of state names, because
//     the set of transitional states is not documented and a missed one
//     leaks a server.
//
//   - Force is required for anything with containers on it, and is
//     harmless otherwise, so it is always set.
//
//   - The reported state lags reality. Servers observed answering SSH
//     were refused a moment later as 'Shutting Down', so a refusal can
//     describe a state the server has already left. Reading GetState on
//     each retry, as below, at least makes the wait legible.
//
//     "Not found" is treated as success rather than as an error, since
//     a server deleted by an earlier attempt — or by anything else —
//     should not be reported as a failure to delete.
//
// Failures are reported per server rather than aborting on the first,
// and the step ends by naming anything it could not remove.
func stepDelete(ctx context.Context, c clients, st *state) error {
	names := st.deletable()
	if len(names) == 0 && st.keyID == "" {
		log.Printf("  nothing to delete")
		return nil
	}

	var leaked []string
	for _, name := range names {
		if err := deleteServer(ctx, c, name); err != nil {
			log.Printf("✗ %v", err)
			leaked = append(leaked, name)
			continue
		}
		log.Printf("✓ deleted %s", name)
	}

	// The ephemeral key is account state too; leaving it behind
	// accumulates clutter on every run.
	if st.keyID != "" {
		time.Sleep(throttle)
		if _, err := c.key.Delete(ctx, sshkey.DeleteRequest{ID: st.keyID}); err != nil {
			log.Printf("✗ ssh/key.Delete(%s): %v", st.keyID, err)
			leaked = append(leaked, "ssh key "+st.keyID)
		} else {
			log.Printf("✓ deleted ephemeral ssh key %s", st.keyID)
		}
	}

	if len(leaked) > 0 {
		return fmt.Errorf("NOT deleted, remove by hand: %s", strings.Join(leaked, ", "))
	}
	return nil
}

// deleteServer deletes one server, waiting out any transitional state.
//
// Rather than guessing from the rejection alone, this reads the server's
// actual state through GetState on each retry and logs it. That turns a
// silent wait into something diagnosable: "still Shutting Down" is a
// normal delay, while a state that never settles is a problem worth
// seeing.
func deleteServer(ctx context.Context, c clients, name string) error {
	var last string
	for attempt := 1; attempt <= deleteRetries; attempt++ {
		time.Sleep(throttle)
		resp, err := c.server.Delete(ctx, server.DeleteRequest{Name: name, Force: true})
		switch {
		case err == nil && resp.Status:
			return nil
		case err != nil:
			last = err.Error()
		default:
			last = resp.Msg
		}

		// A server that has already gone is not a failure to report.
		if strings.Contains(strings.ToLower(last), "not found") {
			log.Printf("  %s is already gone", name)
			return nil
		}
		if !transitionalState(last) {
			return fmt.Errorf("Delete(%s): %s", name, last)
		}

		state := observeState(ctx, c, name)
		log.Printf("  %s is %s; retrying delete (%d/%d)", name, state, attempt, deleteRetries)
		time.Sleep(15 * time.Second)
	}
	return fmt.Errorf("Delete(%s): gave up after %d attempts: %s", name, deleteRetries, last)
}

// observeState reports the server's current state for logging, so a
// wait says what it is waiting on.
func observeState(ctx context.Context, c clients, name string) string {
	time.Sleep(throttle)
	got, err := c.server.GetState(ctx, server.GetStateOptions{Name: name})
	if err != nil {
		return fmt.Sprintf("in an unreadable state (%v)", err)
	}
	if job := got.Return.LastJob; job.ID != "" {
		return fmt.Sprintf("%q (last job %s %s: %s)", got.Return.State, job.Type, job.ID, job.State)
	}
	return fmt.Sprintf("%q", got.Return.State)
}

// transitionalState reports whether a delete was refused because the
// server is mid-transition, and so is worth retrying.
//
// Matched on the message's shape rather than on named states.
// 'Provisioning' and 'Shutting Down' have both been seen; the full set
// is undocumented, and treating an unlisted one as permanent leaks a
// server — which is the more expensive mistake.
func transitionalState(msg string) bool {
	return strings.Contains(strings.ToLower(msg), "cannot be deleted while in the")
}
