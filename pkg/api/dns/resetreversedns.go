package dns

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/models"
	"github.com/sitehostnz/gosh/pkg/net"
)

// ResetReverseDNS clears any custom PTR record for an IP via
// /dns/reset_reverse_dns.json, returning the IP to the platform
// default RDNS.
//
// **Live finding** (May 2026): the API may reject this call with
// "You do not have access to this IP address" even when the same
// client_id can successfully UpdateReverseDNS the same IP — the
// access check is asymmetric between the two endpoints. Workaround:
// call UpdateReverseDNS with the platform default form (e.g.
// "<dashed-ip>.sitehost.co.nz") instead of attempting a reset.
// See docs/api-issues/dns-reset-reverse-dns-asymmetric-access.md.
func (s *Client) ResetReverseDNS(ctx context.Context, request ResetReverseDNSRequest) (response models.APIResponse, err error) {
	u := "dns/reset_reverse_dns.json"
	keys := []string{"client_id", "ip_addr"}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("ip_addr", request.IPAddr)

	req, err := s.client.NewRequest("POST", u, net.Encode(values, keys))
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
