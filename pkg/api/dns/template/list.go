package template

import (
	"context"
)

// List retrieves all DNS domain templates available to the authenticated client.
// It uses the "dns/domain_templates/list_templates.json" API endpoint.
//
// The endpoint is documented at https://docs.sitehost.nz/api/v1.5/?path=/dns
// (the parent /dns index lists it; the per-endpoint reference page is sparse
// at the time of writing — field shapes here mirror what the production API
// returns).
func (s *Client) List(ctx context.Context) (response ListResponse, err error) {
	req, err := s.client.NewRequest("GET", "dns/domain_templates/list_templates.json", "")
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
