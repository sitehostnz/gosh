package srs

import (
	"context"

	"github.com/sitehostnz/gosh/pkg/net"
)

// ListDomains retrieves the list of domains registered for the
// authenticated client via "srs/list_domains.json".
//
// opt may be nil (returns the default page with default sort);
// otherwise SortBy / SortDir / PageSize / PageNumber narrow or
// reorder the result.
//
// The endpoint is documented at https://docs.sitehost.nz/api/v1.5/?path=/srs.
// The per-endpoint reference page is sparse at the time of writing —
// field shapes here mirror the production API's responses.
func (s *Client) ListDomains(ctx context.Context, opt *ListDomainsOptions) (response ListDomainsResponse, err error) {
	u := "srs/list_domains.json"

	path, err := net.AddOptions(u, opt)
	if err != nil {
		return response, err
	}

	req, err := s.client.NewRequest("GET", path, "")
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
