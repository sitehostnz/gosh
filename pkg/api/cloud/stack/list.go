package stack

import (
	"context"
	"fmt"
	"github.com/sitehostnz/gosh/pkg/models"
)

func (s *Client) List(ctx context.Context, request ListRequest) (*[]models.Stack, error) {

	u := fmt.Sprintf("cloud/stack/list_all.json?filters[server_name]=%v", request.ServerName)

	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return nil, err
	}

	response := new(ListResponse)
	err = s.client.Do(ctx, req, response)
	if err != nil {
		return nil, err
	}

	return response.Return.Stacks, nil
}
