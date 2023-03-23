package user

import (
	"context"

	"github.com/sitehostnz/gosh/pkg/utils"
)

// Get fetches a cloud image.
func (s *Client) Get(ctx context.Context, request GetRequest) (response GetResponse, err error) {
	uri := "cloud/ssh/user/get.json"
	keys := []string{
		"apikey",
		"client_id",
		"server_name",
		"username",
	}

	req, err := s.client.NewRequest("GET", uri, "")
	if err != nil {
		return response, err
	}

	values := req.URL.Query()
	values.Add("server_name", request.ServerName)
	values.Add("username", request.Username)

	req.URL.RawQuery = utils.Encode(values, keys)

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
