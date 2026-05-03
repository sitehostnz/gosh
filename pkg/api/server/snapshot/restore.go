package snapshot

import (
	"context"
	"fmt"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Restore restores a server to a previous snapshot via
// "server/snapshot/restore.json". Both Name (the server) and
// Snapshot (the snapshot ID) are required.
//
// **This is destructive.** The server's disk state is reverted
// to the snapshot's contents. Any data written since the
// snapshot was taken will be lost. Treat as a write-mode
// operation requiring deliberate caller intent.
func (s *Client) Restore(ctx context.Context, opt SnapshotOptions) (response JobResponse, err error) {
	if opt.Name == "" {
		return response, fmt.Errorf("snapshot.Restore: Name is required")
	}
	if opt.Snapshot == "" {
		return response, fmt.Errorf("snapshot.Restore: Snapshot is required")
	}

	u := "server/snapshot/restore.json"

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
