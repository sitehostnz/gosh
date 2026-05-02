package volume

import (
	"context"
	"fmt"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Get retrieves the details for a single volume via
// "cloud/volume/get.json". Both Server and Volume are required.
func (s *Client) Get(ctx context.Context, opt GetOptions) (response GetResponse, err error) {
	if opt.Server == "" {
		return response, fmt.Errorf("volume.Get: Server is required")
	}
	if opt.Volume == "" {
		return response, fmt.Errorf("volume.Get: Volume is required")
	}

	u := "cloud/volume/get.json"

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
