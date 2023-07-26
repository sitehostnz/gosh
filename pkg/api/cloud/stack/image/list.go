package image

import (
	"context"
)

// List returns a list of stack images, specific to the customer.
func (s *Client) List(ctx context.Context) (response ListResponse, err error) {
	uri := "cloud/stack/image/list_all.json"

	req, err := s.client.NewRequest("GET", uri, "")
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
