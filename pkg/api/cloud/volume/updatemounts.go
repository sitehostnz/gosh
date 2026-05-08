package volume

import (
	"context"
	"fmt"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/net"
)

// UpdateMounts incrementally adds or removes container mount
// targets for a volume via "cloud/volume/update_mounts.json".
// ServerName and VolumeName are required; at least one of Add
// or Remove must contain a container.
//
// Wire shape:
//
//	containers[add][<stack_name>][] = <container_name>
//	containers[remove][<stack_name>][] = <container_name>
//
// Returned JobResponse carries the scheduler job ID.
func (s *Client) UpdateMounts(ctx context.Context, opt UpdateMountsOptions) (response JobResponse, err error) {
	if opt.ServerName == "" {
		return response, fmt.Errorf("volume.UpdateMounts: ServerName is required")
	}
	if opt.VolumeName == "" {
		return response, fmt.Errorf("volume.UpdateMounts: VolumeName is required")
	}
	if len(opt.Add) == 0 && len(opt.Remove) == 0 {
		return response, fmt.Errorf("volume.UpdateMounts: at least one of Add or Remove must contain a container")
	}

	u := "cloud/volume/update_mounts.json"

	keys := []string{"client_id", "server_name", "volume_name"}
	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("server_name", opt.ServerName)
	values.Add("volume_name", opt.VolumeName)
	for _, c := range opt.Add {
		k := fmt.Sprintf("containers[add][%s][]", c.StackName)
		values.Add(k, c.ContainerName)
		keys = append(keys, k)
	}
	for _, c := range opt.Remove {
		k := fmt.Sprintf("containers[remove][%s][]", c.StackName)
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
