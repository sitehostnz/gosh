package dns_test

import (
	"context"
	"strings"
	"testing"

	"github.com/sitehostnz/gosh/internal/apitest"
	"github.com/sitehostnz/gosh/pkg/api/dns"
)

func TestListZones_DecodesARecordedResponse(t *testing.T) {
	t.Parallel()
	ex := apitest.Serve(t, "list_domains.json")

	got, err := dns.New(ex.Client).ListZones(context.Background(), &dns.ListZoneOptions{})
	if err != nil {
		t.Fatalf("ListZones: %v", err)
	}

	apitest.AssertDecodesFully(t, ex.Body, dns.ListZoneResponse{})

	if len(got.Return.Data) == 0 {
		t.Fatal("no zones decoded; the fixture has one")
	}
	if got.Return.Data[0].Name == "" {
		t.Error("Name is empty; the name is what every other DNS endpoint takes")
	}
	if got.Return.TotalItems == 0 {
		t.Error("TotalItems = 0; the pagination envelope did not decode")
	}
}

// TestGetZone_AbsenceIsNotAnError is the trap this package's docs now
// warn about, pinned so it cannot quietly change.
//
// GetZone is a search. A name matching nothing comes back status:true
// with an empty list, never a rejection — so a caller checking only err
// concludes the zone exists. The fixture is a real response to a name
// under .invalid, which cannot be registered.
func TestGetZone_AbsenceIsNotAnError(t *testing.T) {
	t.Parallel()
	ex := apitest.Serve(t, "search_domains-nomatch.json")

	got, err := dns.New(ex.Client).GetZone(context.Background(),
		dns.GetZoneRequest{DomainName: "sdk-probe-no-such-zone.invalid"})
	if err != nil {
		t.Fatalf("GetZone must not error for a name that matches nothing: %v", err)
	}
	if !got.Status {
		t.Errorf("Status = false, msg = %q; the API reports a successful search", got.Msg)
	}
	if len(got.Return) != 0 {
		t.Fatalf("Return has %d zone(s), want none", len(got.Return))
	}

	// The wire shape it searches on is worth pinning too: this is a
	// POST carrying query[domain], not a GET on a name.
	if ex.Request.Method != "POST" {
		t.Errorf("method = %s, want POST", ex.Request.Method)
	}
}

// TestListRecords_RejectsAnUnknownZone contrasts with the above: this
// endpoint does report absence as an error. The two conventions differ
// within the same package, so neither can be assumed.
//
// # The fixture had to be re-recorded to establish this
//
// The first version was probed with a name under .invalid, and what it
// captured was the API's top-level-domain validation refusing to parse
// the name — "Please specify a valid domain name." — which would have
// come back for a zone under .invalid that *did* exist. It said
// nothing about a well-formed zone that is simply absent, while three
// tests and two doc comments rested on it as though it did.
//
// This fixture is a rejection of a syntactically valid .co.nz name the
// account does not hold, which is the question being asked.
func TestListRecords_RejectsAnUnknownZone(t *testing.T) {
	t.Parallel()
	ex := apitest.Serve(t, "list_records-unknown-zone.json")

	_, err := dns.New(ex.Client).ListRecords(context.Background(),
		dns.ListRecordsRequest{Domain: "gosh-probe-no-such-zone-9f3a2b1c.co.nz"})
	if err == nil {
		t.Fatal("ListRecords: expected an error for a zone that does not exist")
	}
	// The message has to survive into the error, or a caller cannot
	// tell a not-found from any other rejection.
	if !strings.Contains(err.Error(), "doesn't exist") {
		t.Errorf("ListRecords: error is %q, want the API's not-found message", err)
	}
	// Asserted rather than assumed: without a status field the decode
	// would yield false from Go's zero value, so the test would pass
	// against a fixture that recorded nothing.
	apitest.AssertDecodesFully(t, ex.Body, dns.ListRecordsResponse{})
}
