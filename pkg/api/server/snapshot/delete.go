package snapshot

import (
	"context"
	"fmt"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Delete removes a snapshot via "server/snapshot/delete.json".
// Both Name (the server) and Snapshot (the snapshot ID) are
// required.
func (s *Client) Delete(ctx context.Context, opt SnapshotOptions) (response JobResponse, err error) {
	if opt.Name == "" {
		return response, fmt.Errorf("snapshot.Delete: Name is required")
	}
	if opt.Snapshot == "" {
		return response, fmt.Errorf("snapshot.Delete: Snapshot is required")
	}

	u := "server/snapshot/delete.json"

	keys := []string{"client_id", "name", "snapshot"}
	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("name", opt.Name)
	values.Add("snapshot", opt.Snapshot)

	req, err := s.client.NewRequest("POST", u, net.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
