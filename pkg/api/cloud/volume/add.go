package volume

import (
	"context"
	"fmt"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Add creates a new volume on the specified server via
// "cloud/volume/add.json". ServerName and VolumeName are
// required; ContainerNames optionally attaches the volume to
// containers at creation time.
//
// The API queues an asynchronous scheduler job; the returned
// JobResponse carries the job ID for tracking.
func (s *Client) Add(ctx context.Context, opt AddOptions) (response JobResponse, err error) {
	if opt.ServerName == "" {
		return response, fmt.Errorf("volume.Add: ServerName is required")
	}
	if opt.VolumeName == "" {
		return response, fmt.Errorf("volume.Add: VolumeName is required")
	}

	u := "cloud/volume/add.json"

	keys := []string{"client_id", "server_name", "volume_name", "container_names[]"}
	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("server_name", opt.ServerName)
	values.Add("volume_name", opt.VolumeName)
	for _, name := range opt.ContainerNames {
		values.Add("container_names[]", name)
	}

	req, err := s.client.NewRequest("POST", u, net.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
