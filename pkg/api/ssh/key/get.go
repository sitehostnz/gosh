package key

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/utils"
)

// Get returns the specified ssh key.
func (s *Client) Get(ctx context.Context, request GetRequest) (response GetResponse, err error) {
	u := "ssh/key/get.json"

	keys := []string{
		"client_id",
		"key_id",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("key_id", request.ID)

	req, err := s.client.NewRequest("POST", u, utils.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
