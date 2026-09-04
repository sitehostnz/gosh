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
// Verified against a live account, August 2026, with a name under
// .invalid that cannot be registered.
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

	// TODO add control for empty response

	return response, nil
}
