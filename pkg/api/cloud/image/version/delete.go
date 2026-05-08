package version

import (
	"context"
	"fmt"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Delete removes a specific version of a custom image via
// /cloud/image/version/delete.json. Version is the version string
// (e.g. "1.1-1076") as returned by ListAll. Returns a scheduler job.
//
// To delete the *image* itself, use cloud.image.Delete instead —
// this endpoint only prunes one build/version.
func (s *Client) Delete(ctx context.Context, request DeleteRequest) (response JobResponse, err error) {
	if request.Code == "" {
		return response, fmt.Errorf("cloud.image.version.Delete: Code is required")
	}
	if request.Version == "" {
		return response, fmt.Errorf("cloud.image.version.Delete: Version is required")
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("code", request.Code)
	values.Add("version", request.Version)

	req, err := s.client.NewRequest("POST", "cloud/image/version/delete.json",
		net.Encode(values, []string{"client_id", "code", "version"}))
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
