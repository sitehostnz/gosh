package bandwidth

import (
	"context"
)

// ListIpAddresses fetchs a list of IP addresses and subnets
func (s *Client) ListIpAddresses(ctx context.Context) (response ListIpAddressesResponse, err error) {
	u := "bandwidth/get_ip_list.json"
	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
