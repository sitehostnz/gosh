package server

import (
	"context"

	"github.com/sitehostnz/gosh/pkg/net"
)

// GetState retrieves the runtime state of a server (on/off/rescue
// mode and the most recent job affecting it) via
// "server/get_state.json".
func (s *Client) GetState(ctx context.Context, opt GetStateOptions) (response GetStateResponse, err error) {
	u := "server/get_state.json"

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
