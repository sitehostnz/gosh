package stack

import (
	"context"
)

// GenerateName generate a new cloud stack name.
func (s *Client) GenerateName(ctx context.Context) (response GenerateNameResponse, err error) {
	uri := "cloud/stack/generate_name.json"

	req, err := s.client.NewRequest("GET", uri, "")
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, err
}
