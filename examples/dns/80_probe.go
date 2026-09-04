package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/sitehostnz/gosh/pkg/api/dns"
	"github.com/sitehostnz/gosh/pkg/api/dns/template"
)

// probeZone is a zone name that cannot resolve to anything real.
//
// .invalid is reserved by RFC 2606 precisely so it can never be
// registered, which makes every probe below safe: a call naming it is
// rejected before it can do anything, so even the write-shaped
// endpoints have no effect and this step needs no opt-in.
const probeZone = "sdk-probe-no-such-zone.invalid"

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
	probes := append(zoneProbes(), templateProbes()...)

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

// zoneProbes exercise the zone and record endpoints.
func zoneProbes() []probe {
	return []probe{
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
}

// templateProbes exercise the DNS template endpoints, which are a
// separate package and had no coverage at all.
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
// rather than a rejection by it. This API answers HTTP 200 with
// status:false when it refuses, so a rejection arrives decoded;
// anything that never got that far is a transport problem.
func isTransport(err error) bool {
	s := err.Error()
	return strings.Contains(s, "connection refused") ||
		strings.Contains(s, "no such host") ||
		strings.Contains(s, "context deadline exceeded") ||
		strings.Contains(s, "EOF")
}

// oneLine flattens an error for a single log line.
func oneLine(err error) string {
	return strings.Join(strings.Fields(err.Error()), " ")
}
