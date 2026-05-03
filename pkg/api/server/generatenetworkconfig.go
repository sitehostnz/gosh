package server

import (
	"context"

	"github.com/sitehostnz/gosh/pkg/net"
)

// GenerateNetworkConfig retrieves the network configuration files
// for a server (returned as a path → file-contents map) via
// "server/generate_network_config.json".
func (s *Client) GenerateNetworkConfig(ctx context.Context, opt GenerateNetworkConfigOptions) (response GenerateNetworkConfigResponse, err error) {
	u := "server/generate_network_config.json"

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
