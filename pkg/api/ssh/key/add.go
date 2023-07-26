package key

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/utils"
)

// Add adds a new ssh key.
func (s *Client) Add(ctx context.Context, request AddRequest) (response AddResponse, err error) {
	uri := "ssh/key/add.json"
	keys := []string{
		"client_id",
		"label",
		"content",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("label", request.Label)
	values.Add("content", request.Content)

	req, err := s.client.NewRequest("POST", uri, utils.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
