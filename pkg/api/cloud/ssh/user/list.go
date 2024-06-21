package user

import (
	"context"

	"github.com/sitehostnz/gosh/pkg/utils"
)

// List returns a list of stack images, specific to the customer.
func (s *Client) List(ctx context.Context, options ListOptions) (response ListResponse, err error) {
	uri := "cloud/ssh/user/list_all.json"

	path, err := utils.AddOptions(uri, options)
	if err != nil {
		return response, err
	}

	req, err := s.client.NewRequest("GET", path, "")
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
