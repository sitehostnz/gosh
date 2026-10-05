package dns

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/net"
)

// GetZone looks a zone up by name, via "dns/search_domains.json".
//
// # Absence is not an error
//
// This is a search endpoint, not a fetch. A name that matches nothing
// comes back status:true with an empty Return — never a rejection. So
// checking err alone will tell you a zone exists when it does not:
//
//	got, err := c.GetZone(ctx, dns.GetZoneRequest{DomainName: name})
//	if err != nil {
//	    return err
//	}
//	if len(got.Return) == 0 {
//	    return fmt.Errorf("no zone named %s", name)  // <- required
//	}
//
// Being a search also means it can match more than one zone, so the
// first element is not guaranteed to be the name asked for. Compare
// [models.DNSZone.Name] before using it.
//
// Verified against a live account with a well-formed name the account
// does not hold. Note that this endpoint does not validate the
// top-level domain — a name under .invalid returns the same empty
// search rather than a rejection — which is why the evidence is a
// real TLD: the write endpoints do validate, so a .invalid recording
// proves less than it appears to.
func (s *Client) GetZone(ctx context.Context, request GetZoneRequest) (response GetZoneResponse, err error) {
	u := "dns/search_domains.json"

	keys := []string{
		"client_id",
		"query[domain]",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("query[domain]", request.DomainName)

	req, err := s.client.NewRequest("POST", u, net.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	// No empty-response control, deliberately: this endpoint is a
	// search, so a name matching nothing is a legitimate empty result
	// rather than a failure. Callers detect absence with
	// len(Return) == 0 — see the doc comment above.
	//
	// Recorded as a decision rather than left as a gap because it has
	// been re-litigated once already. Adding the control makes
	// TestGetZone_AbsenceIsNotAnError fail, which is the tripwire
	// working, not an oversight to fix.

	return response, nil
}
