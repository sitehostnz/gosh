package stack

import (
	"context"

	"github.com/sitehostnz/gosh/pkg/utils"
)

// Get fetches a cloud stack.
func (s *Client) Get(ctx context.Context, request GetRequest) (response GetResponse, err error) {
	uri := "cloud/stack/get.json"
	keys := []string{
		"apikey",
		"client_id",
		"server",
		"name",
	}

	req, err := s.client.NewRequest("GET", uri, "")
	if err != nil {
		return response, err
	}

	v := req.URL.Query()
	v.Add("server", request.ServerName)
	v.Add("name", request.Name)

	req.URL.RawQuery = utils.Encode(v, keys)

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
