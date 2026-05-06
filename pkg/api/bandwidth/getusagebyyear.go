package bandwidth

import (
	"context"

	"github.com/sitehostnz/gosh/pkg/net"
)

// GetUsageByYear retrieves the annual bandwidth usage for a single
// IP via "bandwidth/get_usage_by_year.json". Period keys in the
// response are "YYYY". The IPAddr field of opt is required and
// must be in CIDR form (e.g. "203.0.113.10/32").
func (s *Client) GetUsageByYear(ctx context.Context, opt UsageOptions) (response UsageResponse, err error) {
	u := "bandwidth/get_usage_by_year.json"

	path, err := net.AddOptions(u, opt)
	if err != nil {
		return response, err
	}

	req, err := s.client.NewRequest("GET", path, "")
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
