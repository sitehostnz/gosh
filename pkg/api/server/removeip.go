package server

import (
	"context"
	"fmt"

	"github.com/google/go-querystring/query"
)

// RemoveIP removes an IP address from a server via
// "server/remove_ip.json". Both Name and IP are required.
// Returns the scheduler job and the IP that was removed.
func (s *Client) RemoveIP(ctx context.Context, opt RemoveIPOptions) (response IPJobResponse, err error) {
	if opt.Name == "" {
		return response, fmt.Errorf("server.RemoveIP: Name is required")
	}
	if opt.IP == "" {
		return response, fmt.Errorf("server.RemoveIP: IP is required")
	}

	values, err := query.Values(opt)
	if err != nil {
		return response, err
	}
	req, err := s.client.NewRequest("POST", "server/remove_ip.json", values.Encode())
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
