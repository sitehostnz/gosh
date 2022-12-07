package server

import (
	"context"
	"fmt"
)

// Get information about the server.
func (s *Client) Get(ctx context.Context, request GetRequest) (response GetResponse, err error) {
	u := fmt.Sprintf("server/get_server.json?name=%v", request.ServerName)

	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
