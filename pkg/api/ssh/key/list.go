package key

import (
	"context"
)

// List returns a list of all ssh keys.
func (s *Client) List(ctx context.Context) (response ListResponse, err error) {
	uri := "ssh/key/list_all.json"
	req, err := s.client.NewRequest("GET", uri, "")
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	// if the request works, don't care about pagination right now.
	return response, err
}
