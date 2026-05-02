package volume

import (
	"context"
	"fmt"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Mount attaches a volume to one or more containers via
// "cloud/volume/mount.json". ServerName, VolumeName, and at
// least one Container are required.
//
// The API expects a nested form-encoded shape:
//
//	containers[<stack_name>][] = <container_name>
//
// repeated for each (StackName, ContainerName) pair. Returned
// JobResponse carries the scheduler job ID.
func (s *Client) Mount(ctx context.Context, opt MountOptions) (response JobResponse, err error) {
	if opt.ServerName == "" {
		return response, fmt.Errorf("volume.Mount: ServerName is required")
	}
	if opt.VolumeName == "" {
		return response, fmt.Errorf("volume.Mount: VolumeName is required")
	}
	if len(opt.Containers) == 0 {
		return response, fmt.Errorf("volume.Mount: at least one Container is required")
	}

	u := "cloud/volume/mount.json"

	keys := []string{"client_id", "server_name", "volume_name"}
	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("server_name", opt.ServerName)
	values.Add("volume_name", opt.VolumeName)
	for _, c := range opt.Containers {
		k := fmt.Sprintf("containers[%s][]", c.StackName)
		values.Add(k, c.ContainerName)
		keys = append(keys, k)
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
