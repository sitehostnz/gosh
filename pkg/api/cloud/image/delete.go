package image

import (
	"context"
	"fmt"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Delete deletes a custom image via /cloud/image/delete.json. The
// API rejects deletion if any container is still using a version of
// this image — clean those up first. Returns a scheduler job.
func (s *Client) Delete(ctx context.Context, request DeleteRequest) (response JobResponse, err error) {
	if request.Code == "" {
		return response, fmt.Errorf("cloud.image.Delete: Code is required")
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("code", request.Code)

	req, err := s.client.NewRequest("POST", "cloud/image/delete.json",
		net.Encode(values, []string{"client_id", "code"}))
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
