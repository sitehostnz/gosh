package info

import (
	"context"
)

// Get information about the API.
func (s *Client) Get(ctx context.Context) (response GetResponse, err error) {
	req, err := s.client.NewRequest("GET", "api/get_info.json", "")
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
