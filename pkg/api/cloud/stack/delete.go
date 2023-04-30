package stack

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/utils"
)

// Delete removes a cloud stack.
func (s *Client) Delete(ctx context.Context, request DeleteRequest) (response ActionResponse, err error) {
	uri := "cloud/stack/delete.json"
	keys := []string{
		"server",
		"name",
	}

	values := url.Values{}
	values.Add("server", request.Server)
	values.Add("name", request.Name)

	req, err := s.client.NewRequest("POST", uri, utils.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
