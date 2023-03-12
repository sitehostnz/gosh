package key

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/utils"
)

// Remove removes the specified key.
func (s *Client) Remove(ctx context.Context, request RemoveRequest) (response RemoveResponse, err error) {
	u := "ssh/key/remove.json"

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
