package server

import (
	"context"
	"fmt"

	"github.com/google/go-querystring/query"
)

// SetPrimaryIP sets the primary IP address for a server via
// "server/set_primary_ip.json". Both Name and IP are required.
// Synchronous (no scheduler job); returns the new primary IP.
func (s *Client) SetPrimaryIP(ctx context.Context, opt SetPrimaryIPOptions) (response SetPrimaryIPResponse, err error) {
	if opt.Name == "" {
		return response, fmt.Errorf("server.SetPrimaryIP: Name is required")
	}
	if opt.IP == "" {
		return response, fmt.Errorf("server.SetPrimaryIP: IP is required")
	}

	values, err := query.Values(opt)
	if err != nil {
		return response, err
	}
	req, err := s.client.NewRequest("POST", "server/set_primary_ip.json", values.Encode())
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
