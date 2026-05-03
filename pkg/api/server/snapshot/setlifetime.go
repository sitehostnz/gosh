package snapshot

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/sitehostnz/gosh/pkg/net"
)

// SetLifetime adjusts the retention period of an existing
// snapshot via "server/snapshot/set_lifetime.json". All three
// fields are required: Name (the server), Snapshot (the
// snapshot ID), Lifetime (hours).
func (s *Client) SetLifetime(ctx context.Context, opt SetLifetimeOptions) (response JobResponse, err error) {
	if opt.Name == "" {
		return response, fmt.Errorf("snapshot.SetLifetime: Name is required")
	}
	if opt.Snapshot == "" {
		return response, fmt.Errorf("snapshot.SetLifetime: Snapshot is required")
	}
	if opt.Lifetime <= 0 {
		return response, fmt.Errorf("snapshot.SetLifetime: Lifetime must be > 0")
	}

	u := "server/snapshot/set_lifetime.json"

	keys := []string{"client_id", "name", "snapshot", "lifetime"}
	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("name", opt.Name)
	values.Add("snapshot", opt.Snapshot)
	values.Add("lifetime", strconv.Itoa(opt.Lifetime))

	req, err := s.client.NewRequest("POST", u, net.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
