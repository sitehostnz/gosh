package volume

import (
	"context"
	"fmt"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Delete removes a volume via "cloud/volume/delete.json". Both
// Server and Volume are required. Note the parameter names
// follow the get convention (server / volume) rather than the
// add convention (server_name / volume_name) — see the
// DeleteOptions doc comment.
func (s *Client) Delete(ctx context.Context, opt DeleteOptions) (response JobResponse, err error) {
	if opt.Server == "" {
		return response, fmt.Errorf("volume.Delete: Server is required")
	}
	if opt.Volume == "" {
		return response, fmt.Errorf("volume.Delete: Volume is required")
	}

	u := "cloud/volume/delete.json"

	keys := []string{"client_id", "server", "volume"}
	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("server", opt.Server)
	values.Add("volume", opt.Volume)

	req, err := s.client.NewRequest("POST", u, net.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
