package snapshot

import (
	"context"
	"fmt"

	"github.com/sitehostnz/gosh/pkg/net"
)

// List retrieves all snapshots for the specified server via
// "server/snapshot/list_all.json". Name is required.
func (s *Client) List(ctx context.Context, opt ListOptions) (response ListResponse, err error) {
	if opt.Name == "" {
		return response, fmt.Errorf("snapshot.List: Name is required")
	}

	u := "server/snapshot/list_all.json"

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
