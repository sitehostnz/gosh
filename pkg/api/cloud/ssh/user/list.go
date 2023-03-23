package user

import (
	"context"
)

// List returns a list of stack images, specific to the customer.
func (s *Client) List(ctx context.Context, options ListOptions) (response ListResponse, err error) {
	uri := "cloud/ssh/user/list_all.json"

	req, err := s.client.NewRequest("GET", uri, "")
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
