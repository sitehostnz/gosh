package server

import (
	"context"
	"fmt"

	"github.com/google/go-querystring/query"
)

// AddIP adds an IP address to a server via "server/add_ip.json".
// Both Name and IP are required. Returns the scheduler job and
// the IP that was added.
func (s *Client) AddIP(ctx context.Context, opt AddIPOptions) (response IPJobResponse, err error) {
	if opt.Name == "" {
		return response, fmt.Errorf("server.AddIP: Name is required")
	}
	if opt.IP == "" {
		return response, fmt.Errorf("server.AddIP: IP is required")
	}

	values, err := query.Values(opt)
	if err != nil {
		return response, err
	}
	req, err := s.client.NewRequest("POST", "server/add_ip.json", values.Encode())
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
