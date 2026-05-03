package server

import (
	"context"

	"github.com/sitehostnz/gosh/pkg/net"
)

// ListIPs returns the IP addresses available for new server
// provisioning at the given location. Use ListLocations to
// enumerate location codes.
func (s *Client) ListIPs(ctx context.Context, opt ListIPsOptions) (response ListIPsResponse, err error) {
	u := "server/list_ips.json"
	keys := []string{"apikey", "client_id", "location"}

	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return response, err
	}
	v := req.URL.Query()
	v.Add("location", opt.Location)
	req.URL.RawQuery = net.Encode(v, keys)

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
