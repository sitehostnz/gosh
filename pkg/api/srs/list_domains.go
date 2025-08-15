package srs

import (
	"context"
)

// ListDomains the registered domain names
func (s *Client) ListDomains(ctx context.Context) (response ListDomainsResponse, err error) {
	u := "srs/list_domains.json"
	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
