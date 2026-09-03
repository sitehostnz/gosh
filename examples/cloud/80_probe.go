package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/sitehostnz/gosh/pkg/api/cloud/db"
	dbuser "github.com/sitehostnz/gosh/pkg/api/cloud/db/user"
	sshuser "github.com/sitehostnz/gosh/pkg/api/cloud/ssh/user"
	"github.com/sitehostnz/gosh/pkg/api/cloud/stack"
)

// probeServer is a name the API cannot resolve to a real server.
//
// Every probe below addresses it, which is what makes them safe to run
// without SH_EXAMPLE_ALLOW_PROVISION: a call that names a server that
// does not exist is rejected before it can do anything, so even the
// write-shaped endpoints have no effect.
const probeServer = "sdk-probe-no-such-server"

// probe is one deliberate call whose rejection we want on record.
type probe struct {
	// what the call is trying to establish.
	what string
	// expect is what we believe the API will say. It is a note, not an
	// assertion — see stepProbe for why probes do not fail the step.
	expect string
	call   func(context.Context, clients) error
}

// stepProbe makes calls we expect the API to reject, and records them.
//
// # Why rejections, and why they get their own step
//
// A test built on a hand-written mock asserts the belief that produced
// the mock, so it can only confirm that belief. Bugs in this SDK have
// survived a green suite exactly that way: a response field declared a
// bool where the API sends a map, so every disk upgrade failed to
// decode while the fixture asserting the wrong shape kept passing.
//
// Recorded rejections are what a mock cannot give you, because you have
// to be wrong on purpose to obtain one. They are also free: nothing
// here is provisioned, so this step is read-only in effect and needs no
// opt-in.
//
// # Probes do not fail the step
//
// A probe that comes back successful is interesting, not broken — it
// means the API accepts something we thought it would not, which is a
// finding rather than an error. The step reports what happened and
// exits zero. Only a transport failure, which means we learned nothing,
// is worth failing on.
// rejectionProbes are calls the API is expected to refuse. They are the
// half a hand-written mock cannot supply, because obtaining one means
// being wrong on purpose.
func rejectionProbes() []probe {
	return []probe{
		{
			what:   "list databases on a server that does not exist",
			expect: "rejected — the filter value is validated, so an empty page never means the filter was ignored",
			call: func(ctx context.Context, c clients) error {
				_, err := c.db.List(ctx, db.ListOptions{ServerName: probeServer})
				return err
			},
		},
		{
			what:   "list databases with no server name at all",
			expect: "ACCEPTED — every database on the account. The filter is genuinely optional",
			call: func(ctx context.Context, c clients) error {
				_, err := c.db.List(ctx, db.ListOptions{})
				return err
			},
		},
		{
			what:   "get a database that does not exist",
			expect: "rejected rather than an empty result",
			call: func(ctx context.Context, c clients) error {
				_, err := c.db.Get(ctx, db.GetRequest{
					ServerName: probeServer, MySQLHost: "mysql", Database: "nosuchdb",
				})
				return err
			},
		},
		{
			what:   "list database users with no server name",
			expect: "ACCEPTED — every database user on the account",
			call: func(ctx context.Context, c clients) error {
				_, err := c.dbUser.List(ctx, dbuser.ListOptions{})
				return err
			},
		},
		{
			what:   "list SSH users with no server name",
			expect: "ACCEPTED — every SSH user on the account",
			call: func(ctx context.Context, c clients) error {
				_, err := c.sshers.List(ctx, sshuser.ListOptions{})
				return err
			},
		},
	}
}

// stackRejectionProbes are the same idea against the stack endpoints,
// split out only to keep each list short enough to read at a glance.
func stackRejectionProbes() []probe {
	return []probe{
		{
			what:   "list stacks on a server that does not exist",
			expect: "rejected — an unknown server is an error here, not an empty list",
			call: func(ctx context.Context, c clients) error {
				_, err := c.stack.List(ctx, stack.ListRequest{ServerName: probeServer})
				return err
			},
		},
		{
			what:   "get a stack that does not exist on a server that does not exist",
			expect: "rejected as an invalid server name — note this endpoint takes server, not server_name",
			call: func(ctx context.Context, c clients) error {
				_, err := c.stack.Get(ctx, stack.GetRequest{
					ServerName: probeServer, Name: "nosuchstack",
				})
				return err
			},
		},
	}
}

// baselineProbes are calls that should succeed. They are the control:
// without one, a page of rejections proves only that the credentials
// are wrong.
func baselineProbes() []probe {
	return []probe{
		{
			what:   "list the stack image catalogue, which takes no arguments",
			expect: "accepted; the baseline that proves a rejection above is about the request",
			call: func(ctx context.Context, c clients) error {
				res, err := c.image.List(ctx)
				if err == nil {
					log.Printf("    %d stack image(s) offered", len(res.Return))
				}
				return err
			},
		},
		{
			what:   "list cloud servers, which also takes no arguments",
			expect: "accepted; note it cannot be filtered by location or label",
			call: func(ctx context.Context, c clients) error {
				res, err := c.cloud.List(ctx)
				if err == nil {
					log.Printf("    %d cloud server(s) on this account", len(res.CloudServers))
				}
				return err
			},
		},
	}
}

func stepProbe(ctx context.Context, c clients, _ *state) error {
	probes := append(append(rejectionProbes(), stackRejectionProbes()...), baselineProbes()...)

	var accepted, rejected, transport int
	for i, p := range probes {
		time.Sleep(throttle)
		log.Printf("  [%d/%d] %s", i+1, len(probes), p.what)
		err := p.call(ctx, c)
		switch {
		case err == nil:
			accepted++
			log.Printf("    accepted — expected: %s", p.expect)
		case isTransport(err):
			transport++
			log.Printf("    ✗ transport failure: %v", err)
		default:
			rejected++
			log.Printf("    rejected: %s", oneLine(err))
		}
	}

	log.Printf("✓ %d probe(s): %d accepted, %d rejected, %d transport failure(s)",
		len(probes), accepted, rejected, transport)
	if transport > 0 {
		return fmt.Errorf("%d probe(s) never reached the API; nothing was learned from them", transport)
	}
	return nil
}

// isTransport reports whether the error is a failure to reach the API
// rather than a rejection by it.
//
// Read from the error tree rather than from the text. Matching
// substrings got this wrong in both directions: it missed real
// transport failures whose wording differs — TLS handshakes, connection
// reset, i/o timeout, network unreachable — and counted them as
// rejections, which is the opposite conclusion and exactly the
// misreading this classification exists to prevent. It also matched
// "EOF", which is short enough to appear inside a rejection message,
// and this API's messages are free text with the request URL embedded.
//
// The structured version separates the two by construction: a rejection
// arrives as *models.ErrorResponse, which is none of these types.
func isTransport(err error) bool {
	var uerr *url.Error
	if errors.As(err, &uerr) {
		return true // never got an API envelope back
	}
	var nerr net.Error
	return errors.As(err, &nerr) || errors.Is(err, context.DeadlineExceeded)
}

// oneLine flattens an error for a single log line.
func oneLine(err error) string {
	return strings.Join(strings.Fields(err.Error()), " ")
}
