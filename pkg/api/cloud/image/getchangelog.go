package image

import (
	"context"
	"fmt"

	"github.com/sitehostnz/gosh/pkg/net"
)

// GetChangelog returns the change log for a *public* SiteHost image
// via /cloud/image/get_changelog.json. This is for SiteHost-provided
// base images (e.g. "sitehost-php55"); custom images don't carry
// platform-managed changelogs.
func (s *Client) GetChangelog(ctx context.Context, request GetChangelogRequest) (response GetChangelogResponse, err error) {
	if request.Code == "" {
		return response, fmt.Errorf("cloud.image.GetChangelog: Code is required")
	}

	keys := []string{"apikey", "client_id", "code"}

	req, err := s.client.NewRequest("GET", "cloud/image/get_changelog.json", "")
	if err != nil {
		return response, err
	}
	v := req.URL.Query()
	v.Add("code", request.Code)
	req.URL.RawQuery = net.Encode(v, keys)

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
