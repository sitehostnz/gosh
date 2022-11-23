package server

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/utils"
)

// Delete deletes a Server with the provided name.
func (s *Client) Delete(ctx context.Context, serverName string) (*DeleteResponse, error) {
	u := "server/delete.json"

	keys := []string{
		"client_id",
		"name",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("name", serverName)

	req, err := s.client.NewRequest("POST", u, utils.Encode(values, keys))
	if err != nil {
		return nil, err
	}

	response := new(DeleteResponse)
	if err := s.client.Do(ctx, req, response); err != nil {
		return nil, err
	}

	return response, nil
}
