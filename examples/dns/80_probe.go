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

	"github.com/sitehostnz/gosh/pkg/api"
	"github.com/sitehostnz/gosh/pkg/api/dns"
	"github.com/sitehostnz/gosh/pkg/api/dns/template"
)

// probeZone is a syntactically valid zone name, in a real TLD, that
// the probe step refuses to write against unless it has established
// the account does not hold it.
//
// # Why a real TLD
//
// The write endpoints validate the top-level domain: CreateZone
// answers "Please specify a valid domain name." for a name under
// .invalid, before any lookup happens. So a .invalid probe against
// those records the validator refusing to parse the name, which is a
// different answer to a different question than "how does this report
// a zone that is merely absent".
//
// That is scoped to the endpoints where it was observed, deliberately.
// It is not a property of the API: dns/search_domains.json answers a
// .invalid name with status:true and an empty list — see
// testdata/search_domains-nomatch.json, which is a recording of
// exactly that. An earlier version of this comment said the API
// validates the TLD, full stop, and the tree contained the
// counterexample at the time it was written.
//
// # Why the safety is a check and not a sentence
//
// Two probes below are write-shaped. While this name was under
// .invalid the validator made them categorically incapable of
// touching anything. A real TLD removes that, and what replaced it was
// an assertion that the account does not hold the name — enforced by
// nothing.
//
// The name is a fixed constant in a public repository, so it is
// guessable by construction, and creating it is precisely what someone
// asking "what does this endpoint say when the zone does exist?" would
// do. DeleteZone takes the zone and its records and the SDK cannot
// undo it. So accountHoldsProbeZone establishes the precondition on
// every run, and the write-shaped probes are omitted when it fails.
//
// The label is fixed rather than random because the recorded
// rejections should be comparable between runs. Scrub replaces
// hostnames in committed fixtures either way, so the benefit is to the
// reader of a raw recording rather than to the fixtures.
const probeZone = "gosh-probe-no-such-zone-9f3a2b1c.co.nz"

// accountHoldsProbeZone reports whether this account holds probeZone.
//
// This is what makes the step safe to run without an opt-in, and it
// has to be a check rather than a sentence.
//
// Two of the probes below are write-shaped: DeleteZone and AddRecord.
// While probeZone was under .invalid the API's own validation made
// them categorically incapable of touching anything, and that was the
// whole guarantee. Moving to a real TLD — necessary, because a
// .invalid probe records the validator rather than the behaviour being
// asked about — removed it and left an assertion in its place: a
// comment saying the account does not hold this name. Nothing enforced
// it.
//
// The name is a fixed constant in a public repository, so it is
// guessable by construction, and creating it is exactly what someone
// investigating "what does this say when the zone does exist?" would
// do. A DeleteZone takes the zone and its records, and nothing in the
// SDK can undo it.
//
// GetZone answers the question directly and is already one of the
// probes, so the cost is one call made earlier than it would have been.
func accountHoldsProbeZone(ctx context.Context, c clients) (bool, error) {
	time.Sleep(throttle)
	got, err := c.dns.GetZone(ctx, dns.GetZoneRequest{DomainName: probeZone})
	if err != nil {
		// Cannot establish the precondition, so the write-shaped
		// probes do not run. Failing here rather than guessing is the
		// point: the alternative is deleting a zone on the strength of
		// a call that did not answer.
		return false, fmt.Errorf("could not establish whether this account holds %s, so the write-shaped probes are unsafe to run: %w", probeZone, err)
	}
	if len(got.Return) > 0 {
		log.Printf("  this account holds %s, so the DeleteZone and AddRecord probes are skipped", probeZone)
		log.Printf("  (they would act on a real zone; the read-only probes still run)")
		return true, nil
	}
	return false, nil
}

// probe is one deliberate call whose outcome we want on record.
type probe struct {
	what   string
	expect string
	call   func(context.Context, clients) error
}

// stepProbe makes calls we expect the API to refuse, and records them.
//
// Recorded rejections are the half a hand-written mock cannot supply,
// because obtaining one means being wrong on purpose. They are also
// the cheapest thing to collect here: nothing is created, so nothing
// has to be cleaned up.
//
// A probe that comes back accepted is a finding, not a failure — it
// means the API allows something we thought it would not. Only a
// transport error, which teaches nothing, fails the step.
func stepProbe(ctx context.Context, c clients, _ *state) error {
	held, err := accountHoldsProbeZone(ctx, c)
	if err != nil {
		return err
	}

	probes := append(zoneProbes(held), templateProbes()...)

	var accepted, rejected, transport, throttled int
	for i, p := range probes {
		time.Sleep(throttle)
		log.Printf("  [%d/%d] %s", i+1, len(probes), p.what)
		err := p.call(ctx, c)
		switch {
		case err == nil:
			accepted++
			log.Printf("    accepted — expected: %s", p.expect)
		case api.IsRateLimited(err):
			// Neither bucket. A throttle reached the API and was
			// refused before the endpoint considered the request, so it
			// teaches nothing about the endpoint — and counting it as a
			// rejection would file it in the corpus as the API's
			// considered answer about an absent zone, which is a false
			// fixture rather than a missing one.
			throttled++
			log.Printf("    throttled, so this probe learned nothing: %s", oneLine(err))
		case isTransport(err):
			transport++
			log.Printf("    ✗ transport failure: %v", err)
		default:
			rejected++
			log.Printf("    rejected: %s", oneLine(err))
		}
	}

	log.Printf("✓ %d probe(s): %d accepted, %d rejected, %d throttled, %d transport failure(s)",
		len(probes), accepted, rejected, throttled, transport)
	if transport > 0 {
		return fmt.Errorf("%d probe(s) never reached the API; nothing was learned from them", transport)
	}
	return nil
}

// zoneProbes exercise the zone and record endpoints.
func zoneProbes(accountHoldsIt bool) []probe {
	probes := []probe{
		{
			what:   "list records for a zone that does not exist",
			expect: "rejected; establishes whether an unknown zone is an error or an empty list",
			call: func(ctx context.Context, c clients) error {
				_, err := c.dns.ListRecords(ctx, dns.ListRecordsRequest{Domain: probeZone})
				return err
			},
		},
		{
			// The trap this journey exists to document. GetZone is a
			// search, so an unknown zone is not an error: it comes back
			// status:true with an empty list. A caller that checks only
			// err will conclude the zone exists.
			what:   "get a zone that does not exist",
			expect: "ACCEPTED with an empty list — GetZone cannot report absence through err",
			call: func(ctx context.Context, c clients) error {
				got, err := c.dns.GetZone(ctx, dns.GetZoneRequest{DomainName: probeZone})
				if err == nil {
					log.Printf("    returned %d zone(s); absence has to be read from the length", len(got.Return))
				}
				return err
			},
		},
		{
			what:   "list zones with no filters at all",
			expect: "accepted; the baseline that proves a rejection above is about the request",
			call: func(ctx context.Context, c clients) error {
				res, err := c.dns.ListZones(ctx, &dns.ListZoneOptions{})
				if err == nil {
					log.Printf("    %d zone(s)", len(res.Return.Data))
				}
				return err
			},
		},
	}

	// Write-shaped, and only safe because accountHoldsProbeZone
	// established the zone is not there. Omitted rather than run-and-
	// hope when it is: a DeleteZone against a real zone takes it and
	// its records, and this step carries no opt-in.
	if accountHoldsIt {
		return probes
	}
	return append(probes, writeShapedZoneProbes()...)
}

// writeShapedZoneProbes are the probes that would act on probeZone if
// it existed. Only reached when it does not.
func writeShapedZoneProbes() []probe {
	return []probe{
		{
			what:   "delete a zone that does not exist",
			expect: "rejected; a delete of nothing should not report success",
			call: func(ctx context.Context, c clients) error {
				_, err := c.dns.DeleteZone(ctx, dns.DeleteZoneRequest{DomainName: probeZone})
				return err
			},
		},
		{
			what:   "add a record to a zone that does not exist",
			expect: "rejected; the zone is checked before the record is validated",
			call: func(ctx context.Context, c clients) error {
				_, err := c.dns.AddRecord(ctx, dns.AddRecordRequest{
					Domain: probeZone, Type: "A", Name: "probe." + probeZone, Content: "203.0.113.1",
				})
				return err
			},
		},
		{
			what:   "add a record with a type that does not exist",
			expect: "rejected, naming whichever it checks first — zone or type",
			call: func(ctx context.Context, c clients) error {
				_, err := c.dns.AddRecord(ctx, dns.AddRecordRequest{
					Domain: probeZone, Type: "NOTAREALTYPE", Name: probeZone, Content: "x",
				})
				return err
			},
		},
	}
}

func templateProbes() []probe {
	return []probe{
		{
			what:   "list DNS templates, which takes no arguments",
			expect: "accepted; a second baseline, for the template package",
			call: func(ctx context.Context, c clients) error {
				res, err := c.template.List(ctx)
				if err == nil {
					log.Printf("    %d template(s)", len(res.Return))
				}
				return err
			},
		},
		{
			// Template 0 is not a probe for absence, as it first
			// appears: it is the real default template, "Manual DNS
			// Settings". Worth knowing before treating 0 as a null id.
			//
			// Note also that the template listing returns system
			// templates (client_id "0") alongside your own, and their
			// domain_count is a platform-wide figure rather than one
			// scoped to your account. Do not report it as yours.
			what:   "get template id 0, which looks invalid and is not",
			expect: "ACCEPTED — 0 is the default template, not an absent one",
			call: func(ctx context.Context, c clients) error {
				got, err := c.template.Get(ctx, template.GetRequest{TemplateID: "0"})
				if err == nil && len(got.Return) > 0 {
					log.Printf("    template 0 exists and is named %q", got.Return[0].TemplateName)
				}
				return err
			},
		},
		{
			// Worth contrasting with GetZone above: the two endpoints
			// disagree about how absence is reported, so a caller
			// cannot assume either convention holds generally.
			//
			// Established against a name the API accepts as
			// well-formed. An earlier version probed a .invalid name
			// and recorded the TLD validator's refusal, which looks
			// like the same evidence and is not.
			what:   "list the records of a template id that does not exist",
			expect: "rejected — unlike GetZone, this one does report absence through an error",
			call: func(ctx context.Context, c clients) error {
				got, err := c.template.ListRecords(ctx, template.ListRecordsRequest{TemplateID: "99999999"})
				if err == nil {
					log.Printf("    returned %d record(s)", len(got.Return))
				}
				return err
			},
		},
	}
}

// isTransport reports whether the error is a failure to reach the API
// rather than a rejection by it.
//
// Read from the error tree rather than from the text. This is the
// implementation from examples/cloud, and the comment travels with it
// because the substring version has now been written twice.
//
// Matching substrings got this wrong in both directions: it missed real
// transport failures whose wording differs — TLS handshakes, connection
// reset, i/o timeout, network unreachable — and counted them as
// rejections, which is the opposite conclusion and exactly the
// misreading this classification exists to prevent. It also matched
// "EOF", which is short enough to appear inside a rejection message,
// and this API's messages are free text with the request URL embedded.
// A rate limit is not in this predicate's remit at all, in either
// version: api.RateLimitError is a plain wrapper, so it is neither a
// *url.Error nor a net.Error and isTransport returns false for it.
// stepProbe handles it with its own case, because a throttle belongs
// in neither the accepted nor the rejected tally.
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
