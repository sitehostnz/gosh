package server

import "context"

// ListAllocatedIPs returns the IP addresses allocated to the
// authenticated client across the SiteHost network, keyed by a
// dotted form of the IP address (IPv4 dots preserved; IPv6 colons
// replaced with dots, double-colon with double-dot).
//
// The endpoint URL is "server/list_allocated_i_ps.json" — note the
// underscore between "i" and "ps". This is the canonical name on
// the API; using "list_allocated_ips" returns "method does not exist."
func (s *Client) ListAllocatedIPs(ctx context.Context) (response ListAllocatedIPsResponse, err error) {
	u := "server/list_allocated_i_ps.json"
	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
