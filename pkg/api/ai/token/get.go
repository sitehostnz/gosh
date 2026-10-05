package token

import (
	"context"
	"errors"
	"net/url"
)

// Get returns a token's details and configuration.
//
// A revoked token is reported as not found, the same as one that never
// existed.
func (s *Client) Get(ctx context.Context, request GetRequest) (response GetResponse, err error) {
	if request.KeyID == "" {
		return response, errors.New("token.Get: KeyID is required")
	}

	req, err := s.client.NewRequest("GET", "ai/token/get.json?key_id="+url.QueryEscape(request.KeyID), "")
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
