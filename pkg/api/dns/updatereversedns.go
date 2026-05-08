package dns

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/models"
	"github.com/sitehostnz/gosh/pkg/net"
)

// UpdateReverseDNS sets the PTR record for an IP address via
// /dns/update_reverse_dns.json. Both the IP and the desired RDNS
// hostname are required.
func (s *Client) UpdateReverseDNS(ctx context.Context, request UpdateReverseDNSRequest) (response models.APIResponse, err error) {
	u := "dns/update_reverse_dns.json"
	keys := []string{"client_id", "ip_addr", "rdns"}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("ip_addr", request.IPAddr)
	values.Add("rdns", request.RDNS)

	req, err := s.client.NewRequest("POST", u, net.Encode(values, keys))
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
