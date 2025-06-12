package firewall

import (
	"context"
	"path"
)

// Get retrieves the server firewall configuration.
func (s *Client) Get(ctx context.Context, request GetRequest) (response GetResponse, err error) {
	uri := path.Join(apiPrefix, "get.json?server=") + request.ServerName

	req, err := s.client.NewRequest("GET", uri, "")
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
