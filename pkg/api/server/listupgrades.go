package server

import (
	"context"

	"github.com/sitehostnz/gosh/pkg/net"
)

// ListUpgrades retrieves the upgrade-availability information for a
// server (current quota usage, available extra-disk pricing, and
// per-slot disk upgrade options) via "server/list_upgrades.json".
func (s *Client) ListUpgrades(ctx context.Context, opt ListUpgradesOptions) (response ListUpgradesResponse, err error) {
	u := "server/list_upgrades.json"

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
