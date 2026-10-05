package token

import (
	"context"
	"errors"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Delete revokes a token. It stops authenticating against the AI
// platform immediately.
//
// Deleting a token that is already revoked fails with "The specified
// AI key could not be found.", so a retried delete reports an error
// even though the token is gone.
func (s *Client) Delete(ctx context.Context, request DeleteRequest) (response DeleteResponse, err error) {
	if request.KeyID == "" {
		return response, errors.New("token.Delete: KeyID is required")
	}

	u := "ai/token/delete.json"

	keys := []string{
		"key_id",
	}

	values := url.Values{}
	values.Add("key_id", request.KeyID)

	req, err := s.client.NewRequest("POST", u, net.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
