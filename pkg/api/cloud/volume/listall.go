package volume

import (
	"context"

	"github.com/sitehostnz/gosh/pkg/net"
)

// List retrieves all volumes for the authenticated client via
// "cloud/volume/list_all.json", with optional filters.
func (s *Client) List(ctx context.Context, opt *ListOptions) (response ListResponse, err error) {
	u := "cloud/volume/list_all.json"

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
