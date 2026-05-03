package dns

import (
	"context"
	"net/url"
	"strconv"

	"github.com/sitehostnz/gosh/pkg/models"
	"github.com/sitehostnz/gosh/pkg/net"
)

// UpdateSOA updates the SOA (Start of Authority) record for a
// hosted zone via /dns/update_soa.json.
//
// All fields are required by the API. NS is the primary nameserver,
// Email is the SOA contact (in @-separated form, not the
// dot-encoded BIND form), and Refresh / Retry / Expire / Minimum
// are the TTL fields in seconds.
func (s *Client) UpdateSOA(ctx context.Context, request UpdateSOARequest) (response models.APIResponse, err error) {
	u := "dns/update_soa.json"
	keys := []string{
		"client_id", "domain", "ns", "email",
		"refresh", "retry", "expire", "minimum",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("domain", request.Domain)
	values.Add("ns", request.NS)
	values.Add("email", request.Email)
	values.Add("refresh", strconv.Itoa(request.Refresh))
	values.Add("retry", strconv.Itoa(request.Retry))
	values.Add("expire", strconv.Itoa(request.Expire))
	values.Add("minimum", strconv.Itoa(request.Minimum))

	req, err := s.client.NewRequest("POST", u, net.Encode(values, keys))
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
