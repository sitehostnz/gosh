package dns

import "context"

// ListIPs retrieves the IP allocations associated with the
// authenticated client via "dns/list_ips.json". The response is a
// map keyed by IP address; each entry includes addressing details
// (netmask, prefix, address family), the allocated server, and rDNS.
func (s *Client) ListIPs(ctx context.Context) (response ListIPsResponse, err error) {
	u := "dns/list_ips.json"

	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
