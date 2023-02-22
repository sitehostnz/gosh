package server

import (
	"context"
)

// List returns a list of stack servers.
func (s *Client) List(ctx context.Context) (response ListResponse, err error) {
	uri := "cloud/server/list_all.json"

	req, err := s.client.NewRequest("GET", uri, "")
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
