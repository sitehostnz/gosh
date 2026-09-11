package dns_test

import (
	"context"
	"encoding/json"
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
// concludes the zone exists.
//
// The fixture is a real response to a well-formed .co.nz name the
// account does not hold. It was previously recorded against a
// .invalid name, which was weaker evidence than it looked: the write
// endpoints reject an unregistrable TLD before doing any lookup, so a
// .invalid recording can capture the validator rather than the search.
// It happens not to here — dns/search_domains.json does not validate
// the TLD — but "happens not to" is the standard this package exists
// to replace.
func TestGetZone_AbsenceIsNotAnError(t *testing.T) {
	t.Parallel()
	ex := apitest.Serve(t, "search_domains-nomatch.json")

	got, err := dns.New(ex.Client).GetZone(context.Background(),
		dns.GetZoneRequest{DomainName: "gosh-probe-no-such-zone-9f3a2b1c.co.nz"})
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
	// Worth making on every fixture, for the reason its own doc gives:
	// it catches a field the API sends that no Go field receives.
	apitest.AssertDecodesFully(t, ex.Body, dns.ListRecordsResponse{})

	// And the other direction, which AssertDecodesFully does not cover:
	// it walks fixture to struct, so a fixture *missing* a key passes
	// it trivially.
	//
	// That matters here specifically. Without a status field the decode
	// yields false from Go's zero value, so this test would report a
	// rejection against a fixture that recorded nothing — which is how
	// the first version of it passed while resting on a recording of
	// the TLD validator. An earlier fix added the call above with a
	// comment claiming it closed this gap; it did not, and the comment
	// was worse than the silence because it stopped anyone re-checking.
	assertFixtureHasStatus(t, ex.Body)
}

// assertFixtureHasStatus fails if the fixture omits the status field.
//
// A rejection this package asserts must be one the API actually
// reported. Go's zero value for bool is false, so a fixture with no
// status key decodes to exactly what a recorded rejection decodes to,
// and every "this is rejected" test would keep passing if a re-record
// dropped the field.
func assertFixtureHasStatus(t *testing.T, body []byte) {
	t.Helper()

	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		t.Fatalf("fixture is not a JSON object: %v", err)
	}
	if _, ok := raw["status"]; !ok {
		t.Error("the fixture has no \"status\" field, so a test asserting a rejection " +
			"passes on Go's zero value rather than on anything the API said")
	}
}
