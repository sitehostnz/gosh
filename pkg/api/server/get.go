package server

import (
	"context"
	"fmt"

	"github.com/sitehostnz/gosh/pkg/models"
)

// Get gets the Server with the provided name.
func (s *ServersService) Get(ctx context.Context, name string) (*models.Server, error) {
	u := fmt.Sprintf("server/get_server.json?name=%v", name)

	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return nil, err
	}

	response := new(models.ServersGetResponse)
	err = s.client.Do(ctx, req, response)
	if err != nil {
		return nil, err
	}

	return &response.Server, nil
}
