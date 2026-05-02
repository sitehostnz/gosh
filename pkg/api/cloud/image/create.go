package image

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Create creates a new custom image via /cloud/image/create.json.
//
// Label is required. Code is optional — the API generates one from
// the label when omitted, but consumers usually want to specify it
// so the image's GitLab repository slug is predictable.
//
// ForkID forks a public SiteHost image (the parent's id, available
// from cloud/image/list_all.json's is_public=1 entries). When zero,
// the image is built from scratch.
//
// SSHKeys grants the listed customer-level SSH key IDs access to
// the backing GitLab repository at gitlab-clients.sitehost.co.nz.
// Without at least one key, the customer cannot push commits.
//
// The endpoint returns a scheduler job; the image record (and its
// GitLab repository) is created asynchronously. Consumers should
// poll cloud/job until the job completes before attempting to clone.
func (s *Client) Create(ctx context.Context, request CreateRequest) (response JobResponse, err error) {
	if request.Label == "" {
		return response, fmt.Errorf("cloud.image.Create: Label is required")
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("label", request.Label)

	keys := []string{"client_id", "label"}

	if request.Code != "" {
		values.Add("params[code]", request.Code)
		keys = append(keys, "params[code]")
	}
	if request.ForkID != 0 {
		values.Add("params[fork_id]", strconv.Itoa(request.ForkID))
		keys = append(keys, "params[fork_id]")
	}
	for i, id := range request.SSHKeys {
		k := fmt.Sprintf("params[ssh_keys][%d]", i)
		values.Add(k, strconv.Itoa(id))
		keys = append(keys, k)
	}

	req, err := s.client.NewRequest("POST", "cloud/image/create.json", net.Encode(values, keys))
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
