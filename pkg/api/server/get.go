package server

import (
	"context"
	"fmt"

	"github.com/sitehostnz/gosh/pkg/models"
)

// Get gets the Server with the provided name.
func (s *Client) Get(ctx context.Context, request GetRequest) (*models.Server, error) {
	u := fmt.Sprintf("server/get_server.json?name=%v", request.ServerName)

	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return nil, err
	}

	response := new(GetResponse)
	if err := s.client.Do(ctx, req, response); err != nil {
		return nil, err
	}

	return &response.Server, nil
}
