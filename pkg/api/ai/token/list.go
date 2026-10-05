package token

import (
	"context"
)

// List returns the account's tokens. Revoked tokens are not included.
func (s *Client) List(ctx context.Context) (response ListResponse, err error) {
	req, err := s.client.NewRequest("GET", "ai/token/list_all.json", "")
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
