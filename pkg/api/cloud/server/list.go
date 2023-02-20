package stackserver

import (
	"context"
	"github.com/sitehostnz/gosh/pkg/models"
)

// List returns a list of stack servers.
func (s *Client) List(ctx context.Context) (*[]models.StackServer, error) {

	req, err := s.client.NewRequest("GET", "cloud/server/list_all.json", "")
	if err != nil {
		return nil, err
	}

	response := new(ListResponse)
	err = s.client.Do(ctx, req, response)
	if err != nil {
		return nil, err
	}

	return response.StackServers, nil
}
