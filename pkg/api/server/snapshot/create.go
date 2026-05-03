package snapshot

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Create takes a new snapshot of a server's disk via
// "server/snapshot/create.json". Name (the server), Partition
// (disk slot, e.g. "scsi0"), and Lifetime (hours) are all required.
// Returned JobResponse carries the scheduler job ID.
func (s *Client) Create(ctx context.Context, opt CreateOptions) (response JobResponse, err error) {
	if opt.Name == "" {
		return response, fmt.Errorf("snapshot.Create: Name is required")
	}
	if opt.Partition == "" {
		return response, fmt.Errorf("snapshot.Create: Partition is required")
	}
	if opt.Lifetime <= 0 {
		return response, fmt.Errorf("snapshot.Create: Lifetime must be > 0")
	}

	u := "server/snapshot/create.json"

	keys := []string{"client_id", "name", "partition", "lifetime"}
	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("name", opt.Name)
	values.Add("partition", opt.Partition)
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
