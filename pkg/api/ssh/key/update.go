package key

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/utils"
)

// Update updates the key with the given ID.
func (s *Client) Update(ctx context.Context, request UpdateRequest) (response UpdateResponse, err error) {
	uri := "ssh/key/update.json"
	keys := []string{
		"client_id",
		"key_id",
		"params[label]",
		"params[content]",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("key_id", request.ID)
	values.Add("params[label]", request.Label)
	values.Add("params[content]", request.Content)

	req, err := s.client.NewRequest("POST", uri, utils.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
