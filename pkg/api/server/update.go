package server

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Update a Server.
func (s *Client) Update(ctx context.Context, opts UpdateRequest) (response UpdateResponse, err error) {
	u := "server/update.json"

	keys := []string{
		"client_id",
		"name",
		"updates[label]",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("name", opts.Name)
	values.Add("updates[label]", opts.Label)

	req, err := s.client.NewRequest("POST", u, net.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
