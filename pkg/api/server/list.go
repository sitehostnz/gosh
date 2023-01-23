package server

import (
	"context"
	"github.com/sitehostnz/gosh/pkg/models"
)

// List information about the server.
func (s *Client) List(ctx context.Context) (*[]models.Server, error) {
	// u := fmt.Sprintf("server/get_server.json?name=%v", request.ServerName)
	u := "server/list_servers.json"
	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return nil, err
	}

	response := new(ListResponse)
	if err := s.client.Do(ctx, req, &response); err != nil {
		return nil, err
	}

	return response.Return.Servers, nil
}
